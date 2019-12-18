package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/allocine/postgresql-streamer-go/services/amqp"
	"github.com/allocine/postgresql-streamer-go/services/postgresql"
	"github.com/allocine/postgresql-streamer-go/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	amqplib "github.com/streadway/amqp"
)

const (
	maxPublicationAttempt = 5
)

var walParserCmdOpts struct {
	configFile      string
	parallelPublish bool
}
var publishCounter int

func createWalParserCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "start",
		Short:             "Create replication slot, then listen and parse incoming data",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: rootCmdPreRun,
		RunE:              walParserCmdE,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			printMemory()
		},
	}
	cmd.PersistentFlags().StringVarP(&walParserCmdOpts.configFile, "file", "f", "", "Configuration file")
	cmd.PersistentFlags().BoolVarP(&walParserCmdOpts.parallelPublish, "parallel-publish", "p", false, "Publish messages in parallel - order will not be preserved")
	//_ = cmd.MarkPersistentFlagRequired("file")

	return cmd
}

func walParserCmdE(cmd *cobra.Command, args []string) error {
	if err := postgresql.InitPostgreSQLSettings(settings.DBDsn(), rootCmdOpts.logLevel >= LogLevelDebug); err != nil {
		return err
	}
	if err := postgresql.CreateSlot(settings.Slot()); err != nil {
		return err
	}
	listenCmd := postgresql.ListenSlotCmd(settings.Slot())
	err := execute(listenCmd)
	fmt.Println(err)

	return nil
}

func amqpConnect() *amqp.Connection {
	return amqp.Connect(
		settings.RabbitMq(),
		settings.Exchange(),
		settings.ExchangeTopic(),
		settings.Queue(),
		settings.RoutingKey(),
		"publisher",
	)
}

func execute(cmd *exec.Cmd) (err error) {
	// Init RabbitMQ
	amqpConnection := amqpConnect()
	err = amqpConnection.ActivateConfirm()
	if err != nil {
		log.Errorf("ActivateConfirm: %s", err)
	}
	registerSignalHandler(cmd, amqpConnection)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("stdoutPipe: %s", err)
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Errorf("stderrPipe: %s", err)
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go process(stdoutPipe, amqpConnection, wg)
	wg.Wait()
	_, errStderr := io.Copy(os.Stderr, stderrPipe)

	err = cmd.Wait()
	if err != nil {
		return
	}

	if errStderr != nil && errStderr != io.EOF {
		err = fmt.Errorf("failed to capture stderr (%s)", errStderr)
	}
	return
}

func process(stdoutPipe io.ReadCloser, amqpConnection *amqp.Connection, wg sync.WaitGroup) {
	log.Debug("> process")
	r := bufio.NewReader(stdoutPipe)
	for {
		buffer, err := r.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Error(err.Error())
			continue
		}
		log.Tracef("Database Event: %s", string(buffer))

		var change postgresql.Wal2Json
		if err := json.Unmarshal(buffer, &change); err != nil {
			errorLogger.WithFields(log.Fields{
				"error":  err,
				"buffer": buffer,
			}).Error("Could not unmarshal buffer")
		} else {
			for _, c := range change.Change {
				if len(change.Timestamp) != 0 {
					c.Timestamp = change.Timestamp
				}
				if walParserCmdOpts.parallelPublish {
					go publish(amqpConnection, c)
				} else {
					publish(amqpConnection, c)
				}
			}
		}
	}
	wg.Done()
	log.Debug("< process")
}

func publish(amqpConnection *amqp.Connection, change *postgresql.Wal2JsonChange) {
	if atomicChange, err := json.Marshal(change); err != nil {
		errorLogger.WithFields(log.Fields{
			"error":  err,
			"change": change,
		}).Error("Could not marshal change")
	} else {
		message := amqplib.Publishing{
			Headers: amqplib.Table{
				"hash": utils.HashBytes(atomicChange),
			},
			ContentType:     "application/json",
			ContentEncoding: "",
			Body:            atomicChange,
			DeliveryMode:    amqplib.Persistent,
			Priority:        0,
		}
		routingKey := fmt.Sprintf(strings.Replace(settings.RoutingKey(), ".#", ".%s.%s", 1), change.Schema, change.Table)

		if err := amqpConnection.PublishWithRetry(maxPublicationAttempt, true, settings.Exchange(), routingKey, message); err != nil {
			errorLogger.WithFields(log.Fields{
				"error":  err,
				"change": change,
			}).Error("Give up publishing message")
		} else {
			publishCounter++
			log.Debugf("Publish message (counter: %d)", publishCounter)
		}
	}
}

func registerSignalHandler(cmd *exec.Cmd, rabbitmq *amqp.Connection) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-signalChan
		log.Printf("Received SIGNAL %s. Exiting.", sig)

		if err := cmd.Process.Signal(sig); err != nil {
			log.Errorf("Cannot send signal %v to PostgreSQL process", sig)
			if err := cmd.Process.Kill(); err != nil {
				log.Errorf("Cannot send signal %v to PostgreSQL process", os.Kill)
			}
		}

		rabbitmq.Disconnect()
		os.Exit(0)
	}()
}

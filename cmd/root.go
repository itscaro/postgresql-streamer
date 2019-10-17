// Copyright (c) 2019.
// Author: Quan TRAN

package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/allocine/postgresql-streamer-go/services/postgresql"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmdOpts struct {
	noANSI       bool
	logLevel     int
	logFile      string
	errorLogFile string
	dev          bool
	binDir       string
}

var settings *Settings
var errorLogger *log.Logger

const (
	LogLevelNormal      int = 0
	LogLevelVerbose     int = 1
	LogLevelVeryVerbose int = 2
	LogLevelDebug       int = 3
)

func createRootCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:               "",
		Short:             "",
		SilenceUsage:      true,
		SilenceErrors:     true,
		PersistentPreRunE: rootCmdPreRun,
		RunE:              helpCmd,
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			printMemory()
		},
	}

	cmd.PersistentFlags().BoolVar(&rootCmdOpts.noANSI, "no-ansi", false, "Do not use ANSI color")
	cmd.PersistentFlags().IntVarP(&rootCmdOpts.logLevel, "logLevel", "l", LogLevelNormal, "log level (0-3)")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.logFile, "logfile", "-", "log file")
	cmd.PersistentFlags().StringVar(&rootCmdOpts.errorLogFile, "errorLogfile", "-", "Error log file, logs event which cannot be published")

	cmd.AddCommand(
		createWalCmd(),
		createDevCmd(),
	)

	return cmd
}

func rootCmdPreRun(cmd *cobra.Command, args []string) error {
	if s, err := InitSettings(); err != nil {
		return err
	} else {
		settings = s
	}
	postgresql.InitWal2JsonSettings()

	// log level
	if !cmd.Flag("logLevel").Changed && "1" == os.Getenv("APP_DEBUG") {
		rootCmdOpts.logLevel = 3
	}

	// errorLogger is used to store failed publications
	errorLogger = log.New()
	errorLogger.SetFormatter(&log.JSONFormatter{})
	if rootCmdOpts.errorLogFile == "-" {
		errorLogger.SetOutput(os.Stderr)
	} else {
		if f, err := os.OpenFile(rootCmdOpts.errorLogFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600); err != nil {
			log.Fatalf("Could not open log file: %s", err)
		} else {
			errorLogger.SetOutput(f)
		}
	}

	// normalLogger
	if rootCmdOpts.logLevel >= LogLevelDebug {
		log.SetReportCaller(true)
		log.SetLevel(log.TraceLevel)
	} else if rootCmdOpts.logLevel >= LogLevelVeryVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		DisableColors: rootCmdOpts.noANSI,
		FullTimestamp: false,
	})
	if rootCmdOpts.logFile == "-" {
		log.SetOutput(os.Stdout)
	} else {
		if f, err := os.OpenFile(rootCmdOpts.logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600); err != nil {
			log.Fatalf("Could not open log file: %s", err)
		} else {
			log.SetOutput(f)
		}
	}

	return nil
}

func Execute() {
	if err := createRootCmd().Execute(); err != nil {
		log.Errorf("%s", err)
		os.Exit(1)
	}
}

func helpCmd(cmd *cobra.Command, args []string) error {
	return cmd.Help()
}

func printMemory() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Memory (total allocation) %.2f MB\n", float64(m.TotalAlloc)/1000000)
}

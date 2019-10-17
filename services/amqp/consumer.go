package amqp

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Consume message with auto ack. If failure in process, the message is lost
func (c *Connection) ConsumeAutoAck(controlChan chan<- bool, consumerTag string, f func(*Connection, []byte) error) {
	c.consume(controlChan, consumerTag, true, f)
}

// Consume message without auto ack
func (c *Connection) Consume(controlChan chan<- bool, consumerTag string, f func(*Connection, []byte) error) {
	c.consume(controlChan, consumerTag, false, f)
}

func (c *Connection) consume(controlChan chan<- bool, consumerTag string, autoAck bool, f func(*Connection, []byte) error) {
	msgs, err := c.channel.Consume(
		c.queue.Name,
		consumerTag,
		autoAck,
		false,
		false,
		false,
		nil,
	)
	c.logFatal(err, "Failed to register a consumer")

	go func() {
		emptyMessageCounter := 0
		isTerminated := false
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

		for {
			select {
			case d, _ := <-msgs:
				if len(d.Body) == 0 {
					emptyMessageCounter++
					c.log("Received an empty message, wait for %d second(s)", emptyMessageCounter)
					time.Sleep(time.Duration(emptyMessageCounter) * time.Second)
					if emptyMessageCounter > 100 {
						c.log("Received too many empty message consecutively, stopping the worker")
						// If there are too much empty messages, stop the consumer
						isTerminated = true
						controlChan <- true
					}
				} else {
					c.log("Received a message: %s", d.Body)
					emptyMessageCounter = 0
					err := f(c, d.Body)

					if autoAck == false {
						var errAck error
						if err == nil {
							errAck = d.Ack(false)
						} else {
							time.Sleep(1 * time.Second)
							errAck = d.Nack(false, true)
						}
						if errAck != nil {
							c.log("Could ACK/NACK the message")
						}
					}
				}
			case sig := <-signalChan:
				c.log("Received SIGNAL %v", sig)
				isTerminated = true
				controlChan <- true
			}

			if isTerminated {
				break
			}
		}
	}()

	c.log("Consumer %s is waiting for messages", consumerTag)
}

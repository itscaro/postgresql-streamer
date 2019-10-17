// Copyright (c) 2019.
// Author: Quan TRAN

package amqp

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/streadway/amqp"
)

type Connection struct {
	id                string
	url               string
	exchange          string
	routingKey        string
	mutexReconnection *sync.Mutex
	connection        *amqp.Connection
	channel           *amqp.Channel
	queue             *amqp.Queue
}

// Connect creates a connection
// exchange, queue declaration & binding are handled in this function
func Connect(url string, exchange string, exchangeType string, queue string, routingKey string, id string) *Connection {
	if len(url) == 0 {
		log.Fatalf("url cannot be empty")
	}
	if len(queue) == 0 {
		log.Fatalf("queue cannot be empty")
	}
	if len(id) == 0 {
		key := [4]byte{}
		randomID := ""
		if _, err := rand.Read(key[:]); err == nil {
			randomID = hex.EncodeToString(key[:])
		}
		id = fmt.Sprintf("%s-%s-%s", exchange, queue, randomID)
	}

	c := Connection{
		id:                id,
		url:               url,
		exchange:          exchange,
		routingKey:        routingKey,
		mutexReconnection: &sync.Mutex{},
	}

	var err error

	c.connection, err = amqp.Dial(c.url)
	c.logFatal(err, "Failed to connect to RabbitMQ")

	c.channel, err = c.connection.Channel()
	c.logFatal(err, "Failed to open a channel")

	c.initExchange(exchangeType)
	c.initQueue(queue)

	go c.reconnectConnection()
	go c.reconnectChannel()

	return &c
}

func (c *Connection) reconnectConnection() {
	for {
		counter := 0
		reason, ok := <-c.connection.NotifyClose(make(chan *amqp.Error))
		if !ok {
			c.logDebug("connection closed")
			c.connection.Close()
			break
		}
		c.logDebug("connection closed, reason: %v", reason)

		c.mutexReconnection.Lock()
		for {
			counter++
			sleep(counter, 10, time.Second)
			conn, err := amqp.Dial(c.url)
			if err == nil {
				c.connection = conn
				c.logDebug("reconnect success")
				break
			} else if counter == 1000 {
				c.logFatal(err, "tried for 1000 times but could not create new connection")
			}
			c.logDebug("reconnection failed (iteration: %d), err: %v", counter, err)
		}
		c.mutexReconnection.Unlock()
	}
}

func (c *Connection) reconnectChannel() {
	for {
		counter := 0
		reason, ok := <-c.channel.NotifyClose(make(chan *amqp.Error))
		if !ok {
			c.logDebug("channel closed")
			c.channel.Close()
			break
		}
		c.logDebug("channel closed, reason: %v", reason)

		c.mutexReconnection.Lock()
		for {
			counter++
			sleep(counter, 10, time.Second)
			ch, err := c.connection.Channel()
			if err == nil {
				c.logDebug("channel recreate success")
				c.channel = ch
				break
			} else if counter == 1000 {
				c.logFatal(err, "tried for 1000 times but could not recreate new channel")
			}
			c.logDebug("channel recreation failed (iteration: %d), err: %v", counter, err)
		}
		c.mutexReconnection.Unlock()
	}
}

func sleep(sleep int, maxSleep int, unit time.Duration) {
	if sleep > maxSleep {
		sleep = maxSleep
	}
	time.Sleep(time.Duration(sleep) * time.Second)
}

func (c *Connection) Disconnect() {
	c.channel.Close()
	c.connection.Close()
}

func (c *Connection) log(format string, args ...interface{}) {
	args = append([]interface{}{c.queue.Name, c.id}, args...)
	log.Infof("[%s] [%s] "+format, args...)
}

func (c *Connection) logDebug(format string, args ...interface{}) {
	args = append([]interface{}{c.queue.Name, c.id}, args...)
	log.Debugf("[%s] [%s] "+format, args...)
}

func (c *Connection) logTrace(format string, args ...interface{}) {
	args = append([]interface{}{c.queue.Name, c.id}, args...)
	log.Tracef("[%s] [%s] "+format, args...)
}

func (c *Connection) logError(format string, args ...interface{}) {
	args = append([]interface{}{c.queue.Name, c.id}, args...)
	log.Errorf("[%s] [%s] "+format, args...)
}

func (c *Connection) logFatal(err error, msg string) {
	if err != nil {
		if c.queue == nil {
			log.Fatalf("[%s] %s: %s", c.id, msg, err)
		} else {
			log.Fatalf("[%s] [%s] %s: %s", c.queue.Name, c.id, msg, err)
		}
	}
}

func (c *Connection) initExchange(exchangeType string) {
	err := c.channel.ExchangeDeclare(
		c.exchange,
		exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	c.logFatal(err, "Failed to declare an exchange")
}

func (c *Connection) initQueue(queue string) {
	q, err := c.channel.QueueDeclare(
		queue,
		true,
		false,
		false,
		false,
		nil,
	)
	c.queue = &q
	c.logFatal(err, "Failed to declare a queue")

	err = c.channel.QueueBind(
		queue,
		c.routingKey,
		c.exchange,
		false,
		nil,
	)
	c.logFatal(err, "Failed to bind a queue")
}

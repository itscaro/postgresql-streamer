package amqp

import (
	"time"

	"github.com/streadway/amqp"
)

func (c *Connection) Publish(reliable bool, exchange, routingKey string, msg amqp.Publishing) error {
	if reliable {
		// TODO debug deadlock
		//confirmChan := c.channel.NotifyPublish(make(chan amqp.Confirmation, 10))
		//defer c.confirmOne(confirmChan)
	}

	c.mutexReconnection.Lock()
	c.logTrace("Publish message")
	if err := c.channel.Publish(exchange, routingKey, false, false, msg); err != nil {
		c.logDebug("Publication failed")
		return err
	}
	c.mutexReconnection.Unlock()
	return nil
}

func (c *Connection) PublishWithRetry(maxAttempt int, reliable bool, exchange, routingKey string, msg amqp.Publishing) error {
	counter := 0
Publish:
	if err := c.Publish(true, exchange, routingKey, msg); err != nil {
		counter++
		if counter < maxAttempt {
			c.mutexReconnection.Lock()
			c.mutexReconnection.Unlock()
			c.logError("%s. Retry publication after %d second(s)", err, counter)
			time.Sleep(time.Duration(counter) * time.Second)
			goto Publish
		} else {
			c.logError("%s. Aborting publication after %d retries", err, counter)
			return err
		}
	}

	return nil
}

//func (c *Connection) confirmOne(confirms chan amqp.Confirmation) {
//	c.logTrace("waiting for confirmation of one publishing")
//	if confirmed := <-confirms; confirmed.Ack {
//		c.logTrace("confirmed delivery with delivery tag: %d", confirmed.DeliveryTag)
//	} else {
//		c.logTrace("failed delivery of delivery tag: %d", confirmed.DeliveryTag)
//	}
//}

func (c *Connection) ActivateConfirm() error {
	c.log("Enabling publishing confirms")
	if err := c.channel.Confirm(false); err != nil {
		c.logFatal(err, "Channel could not be put into confirm mode")
		return err
	}
	return nil
}

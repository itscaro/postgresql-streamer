package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/streadway/amqp"
)

type Settings struct {
	projectName      string
	slotTables       string
	rabbitmqUri      string
	elasticSearchUri string
	postgresqlDsn    string
}

func InitSettings() (*Settings, error) {
	var settings Settings

	settings.projectName = os.Getenv("PGS_PROJECT_NAME")
	settings.rabbitmqUri = os.Getenv("PGS_RABBITMQ_URI")
	settings.postgresqlDsn = os.Getenv("PGS_PG_DSN")

	if len(settings.projectName) == 0 {
		return nil, errors.New("PGS_PROJECT_NAME must be given")
	}
	if len(settings.rabbitmqUri) == 0 {
		return nil, errors.New("PGS_RABBITMQ_URI must be given")
	}
	if len(settings.postgresqlDsn) == 0 {
		return nil, errors.New("PGS_PG_DSN must be given")
	}

	return &settings, nil
}

func (s *Settings) RabbitMq() string {
	return settings.rabbitmqUri
}

func (s *Settings) RoutingKey() string {
	return fmt.Sprintf("%s_event.#", s.projectName)
}

func (s *Settings) Exchange() string {
	return fmt.Sprintf("%s_event_exchange", s.projectName)
}

func (s *Settings) ExchangeTopic() string {
	return amqp.ExchangeTopic
}

func (s *Settings) Queue() string {
	return fmt.Sprintf("%s_event", s.projectName)
}

func (s *Settings) Slot() string {
	return fmt.Sprintf("%s_wal_parser", s.projectName)
}

func (s *Settings) SlotTables() string {
	return settings.slotTables
}

func (s *Settings) DBDsn() string {
	return settings.postgresqlDsn
}

//TODO remove
type MqSettings struct {
	Url    string            `yaml:"uri"`
	Queues []MqQueueSettings `yaml:"queues"`
}

type MqQueueSettings struct {
	Exchange     string `yaml:"exchange"`
	ExchangeType string `yaml:"exchange_type"`
	Queue        string `yaml:"queue_name"`
	RoutingKey   string `yaml:"routing_key"`
}

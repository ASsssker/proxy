package config

import (
	"net"
	"net/url"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	RequesterServiceConfig
	RabbitMQConfig
}

type RequesterServiceConfig struct {
	RequesterWorkersCount uint `env:"REQUESTER_WORKERS_COUNT"`
	// TODO: coming soon
}

type RabbitMQConfig struct {
	RabbitHost          string `env:"RABBIT_HOST"`
	RabbitPort          string `env:"RABBIT_PORT"`
	RabbitUser          string `env:"RABBIT_USER"`
	RabbitPassword      string `env:"RABBIT_PASSWORD"`
	RabbitTaskQueueName string `env:"RABBIT_TASK_QUEUE_NAME"`
}

func (r RabbitMQConfig) RabbitDNS() string {
	u := url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(url.QueryEscape(r.RabbitUser), url.QueryEscape(r.RabbitPassword)),
		Host:   net.JoinHostPort(r.RabbitHost, r.RabbitPort),
		Path:   "/",
	}

	return u.String()
}

func MustLoad() Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	return cfg
}

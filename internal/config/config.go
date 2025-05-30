package config

import (
	"net"
	"net/url"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env string `env:"ENV"`
	ProxyServiceConfig
	RequesterServiceConfig
	PostgresConfig
	RabbitMQConfig
	NatsMQCOnfig
}

type ProxyServiceConfig struct {
	ProxyHost             string        `env:"PROXY_HOST"`
	ProxyPort             string        `env:"PROXY_PORT"`
	ProxyHTTPReadTimeout  time.Duration `env:"PROXY_HTTP_READ_TIMEOUT"`
	ProxyHTTPWriteTimeout time.Duration `env:"PROXY_HTTP_WRITE_TIMEOUT"`
	ProxyHTTPIdleTimeout  time.Duration `env:"PROXY_HTTP_IDLE_TIMEOUT"`
}

type RequesterServiceConfig struct {
	RequesterWorkersCount      uint          `env:"REQUESTER_WORKERS_COUNT"`
	RequesterHTTPClientTimeout time.Duration `env:"REQUESTER_HTTP_CLIENT_TIMEOUT"`
	RequesterRetryCount        uint          `env:"REQUESTER_RETRY_COUNT"`
	RequesterMetricHost        string        `env:"REQUESTER_METRIC_HOST"`
	RequesterMetricPort        string        `env:"REQUESTER_METRIC_PORT"`
}

type PostgresConfig struct {
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresDB       string `env:"POSTGRES_DB"`
	PostgresPort     string `env:"POSTGRES_PORT"`
	PostgresHost     string `env:"POSTGRES_HOST"`
}

func (p PostgresConfig) PostgresDNS() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(p.PostgresUser, p.PostgresPassword),
		Host:     net.JoinHostPort(p.PostgresHost, p.PostgresPort),
		Path:     "/" + p.PostgresDB,
		RawQuery: buildQuery(map[string]string{"sslmode": "disable"}),
	}

	return u.String()
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
		User:   url.UserPassword(r.RabbitUser, r.RabbitPassword),
		Host:   net.JoinHostPort(r.RabbitHost, r.RabbitPort),
		Path:   "/",
	}

	return u.String()
}

type NatsMQCOnfig struct {
	NatsHost          string `env:"NATS_HOST"`
	NatsPort          string `env:"NATS_PORT"`
	NatsTaskQueueName string `env:"NATS_TASK_QUEUE_NAME"`
}

func (n NatsMQCOnfig) NatsDNS() string {
	u := url.URL{
		Scheme: "nats",
		Host:   net.JoinHostPort(n.NatsHost, n.NatsPort),
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

func buildQuery(params map[string]string) string {
	q := url.Values{}
	for k, v := range params {
		q.Add(k, v)
	}
	return q.Encode()
}

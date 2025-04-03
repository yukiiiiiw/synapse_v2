package config

import (
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	commoncfg "synapse/common/config"
)

var (
	Config *config
	DB     *gorm.DB
	Redis  redis.UniversalClient
)

type config struct {
	App            App                        `json:"app" yaml:"app" mapstructure:"app"`
	Logger         commoncfg.LoggerConfig     `json:"logger" yaml:"logger" mapstructure:"logger"`
	Server         commoncfg.ServerConfig     `json:"server" yaml:"server" mapstructure:"server"`
	Datasource     commoncfg.DatasourceConfig `json:"datasource" yaml:"datasource" mapstructure:"datasource"`
	Redis          commoncfg.RedisConfig      `json:"redis" yaml:"redis" mapstructure:"redis"`
	Sentry         commoncfg.SentryConfig     `json:"sentry" yaml:"sentry" mapstructure:"sentry"`
	AgentNode      AgentNodeConfig            `json:"agent_node" yaml:"agent_node" mapstructure:"agent_node"`
	RabbitMQConfig RabbitMQConfig             `json:"rabbitmq" yaml:"rabbitmq" mapstructure:"rabbitmq"`
}

type App struct {
	AsyncApiWaitTimeout time.Duration   `json:"async_api_wait_timeout" yaml:"async_api_wait_timeout" mapstructure:"async_api_wait_timeout"`
	AuthToken           string          `json:"auth_token" yaml:"auth_token" mapstructure:"auth_token"`
	Services            []ServiceConfig `json:"services" yaml:"services" mapstructure:"services"`
}

type ServiceConfig struct {
	Name     string `json:"name" yaml:"name" mapstructure:"name"`
	Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
}

type AgentNodeConfig struct {
	BusyThreshold    float64       `yaml:"busyThreshold" json:"busyThreshold"`
	HeartbeatTimeout time.Duration `yaml:"heartbeatTimeout" json:"heartbeatTimeout"`
	ScheduleTimeout  time.Duration `yaml:"scheduleTimeout" json:"scheduleTimeout"`
}

type RabbitMQConfig struct {
	Amqp      AmqpMQ           `yaml:"amqp"`
	Producer  ProducerConfig   `yaml:"producer"`
	Consumers []ConsumerConfig `yaml:"consumers"`
}

type AmqpMQ struct {
	Uri  string     `yaml:"uri"`
	Pool PoolConfig `yaml:"pool"`
}

type PoolConfig struct {
	MaxConnections     int           `json:"maxConnections" yaml:"maxConnections"`
	MaxChannelsPerConn int           `json:"maxChannelsPerConn" yaml:"maxChannelsPerConn"`
	WaitTimeout        time.Duration `json:"waitTimeout" yaml:"waitTimeout"`
	ReconnectInterval  time.Duration `json:"reconnectInterval" yaml:"reconnectInterval"`
}

type ProducerConfig struct {
	Name           string           `json:"name" yaml:"name"`
	Retries        int              `json:"retries" yaml:"retries"`
	RetryInterval  time.Duration    `json:"retryInterval" yaml:"retryInterval"`
	ExchangeConfig []ExchangeConfig `json:"exchanges" yaml:"exchanges" mapstructure:"exchanges"`
}
type ExchangeConfig struct {
	Name    string `json:"name" yaml:"name"`
	Type    string `json:"type" yaml:"type"`
	Durable bool   `json:"durable" yaml:"durable"`
}
type ConsumerConfig struct {
	Name          string `json:"name" yaml:"name"`
	Queue         string `json:"queue" yaml:"queue"`
	Exchange      string `json:"exchange" yaml:"exchange"`
	RoutingKey    string `json:"routingKey" yaml:"routingKey"`
	ExchangeType  string `json:"exchangeType" yaml:"exchangeType"`
	Durable       bool   `json:"durable" yaml:"durable"`
	AutoAck       bool   `json:"autoAck" yaml:"autoAck"`
	PrefetchCount int    `json:"prefetchCount" yaml:"prefetchCount"`
}

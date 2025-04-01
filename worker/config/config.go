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
	App             App                        `json:"app" yaml:"app" mapstructure:"app"`
	Logger          commoncfg.LoggerConfig     `json:"logger" yaml:"logger" mapstructure:"logger"`
	Server          commoncfg.ServerConfig     `json:"server" yaml:"server" mapstructure:"server"`
	Datasource      commoncfg.DatasourceConfig `json:"datasource" yaml:"datasource" mapstructure:"datasource"`
	Redis           commoncfg.RedisConfig      `json:"redis" yaml:"redis" mapstructure:"redis"`
	Sentry          commoncfg.SentryConfig     `json:"sentry" yaml:"sentry" mapstructure:"sentry"`
	AgentNodeConfig AgentNodeConfig            `json:"agentNode" yaml:"agentNode" mapstructure:"agentNode"`
}

type App struct {
	AsyncApiWaitTimeout time.Duration   `json:"async_api_wait_timeout" yaml:"async_api_wait_timeout" mapstructure:"async_api_wait_timeout"`
	AuthToken           string          `json:"auth_token" yaml:"auth_token" mapstructure:"auth_token"`
	Services            []ServiceConfig `json:"services" yaml:"services" mapstructure:"services"`
	AgentNode           AgentNodeConfig `json:"agent_node" yaml:"agent_node" mapstructure:"agent_node"`
}

type ServiceConfig struct {
	Name     string `json:"name" yaml:"name" mapstructure:"name"`
	Endpoint string `json:"endpoint" yaml:"endpoint" mapstructure:"endpoint"`
}

type AgentNodeConfig struct {
	BusyThreshold float64 `yaml:"busyThreshold" json:"busyThreshold"`
}

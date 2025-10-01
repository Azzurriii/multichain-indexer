package config

import (
	"time"

	"github.com/fystack/multichain-indexer/internal/rpc"
	"github.com/fystack/multichain-indexer/pkg/common/enum"
)

type Env string

const (
	DevEnv  Env = "development"
	ProdEnv Env = "production"
)

type Config struct {
	Version     string   `yaml:"version"`
	Environment Env      `yaml:"env"      validate:"required,oneof=development production"`
	Defaults    Defaults `yaml:"defaults" validate:"required"`
	Chains      Chains   `yaml:"chains"   validate:"required,min=1"`
	Services    Services `yaml:"services" validate:"required"`
}

type Defaults struct {
	FromLatest          bool               `yaml:"from_latest"`
	PollInterval        time.Duration      `yaml:"poll_interval"         validate:"required"`
	ReorgRollbackWindow int                `yaml:"reorg_rollback_window" validate:"required,min=1"`
	Client              ClientConfig       `yaml:"client"`
	Throttle            Throttle           `yaml:"throttle"`
	Failover            rpc.FailoverConfig `yaml:"failover"`
}

type Chains map[string]ChainConfig

type ChainConfig struct {
	Name                string           `yaml:"-"`
	NetworkId           string           `yaml:"network_id"`
	InternalCode        string           `yaml:"internal_code"`
	Type                enum.NetworkType `yaml:"type"                  validate:"required,oneof=tron evm btc"`
	FromLatest          bool             `yaml:"from_latest"`
	StartBlock          int              `yaml:"start_block"           validate:"min=0"`
	PollInterval        time.Duration    `yaml:"poll_interval"`
	ReorgRollbackWindow int              `yaml:"reorg_rollback_window"`
	Client              ClientConfig     `yaml:"client"`
	Throttle            Throttle         `yaml:"throttle"`
	Nodes               []NodeConfig     `yaml:"nodes"                 validate:"required,min=1"`
}

type ClientConfig struct {
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries" validate:"min=0"`
	RetryDelay time.Duration `yaml:"retry_delay"`
}

type Throttle struct {
	RPS         int `yaml:"rps"`
	Burst       int `yaml:"burst"`
	BatchSize   int `yaml:"batch_size"`
	Concurrency int `yaml:"concurrency"`
}

type NodeConfig struct {
	URL  string     `yaml:"url"  validate:"required,url"`
	Auth AuthConfig `yaml:"auth"`
}

type AuthConfig struct {
	Type  string `yaml:"type"  validate:"omitempty,oneof=header query"`
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

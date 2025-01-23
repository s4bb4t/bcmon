package config

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

const (
	stageDev   = "dev"
	stageProd  = "prod"
	stageLocal = "local"
)

type Config struct {
	Db struct {
		Postgres Postgres `json:"postgres" mapstructure:"postgres"`
	} `json:"db" mapstructure:"db"`

	Preload struct {
		InputData []string `json:"input_data"`
	}

	Debug bool

	Network     string `mapstructure:"network" json:"network"`
	UpstreamURL string `mapstructure:"upstream_url" json:"upstream_url"`

	SubgraphPath string `mapstructure:"subgraph_path" json:"subgraph_path"`

	RequestDelay time.Duration `mapstructure:"request_delay" json:"request_delay"`
	UpdateDelay  time.Duration `mapstructure:"update_delay" json:"update_delay"`
}

func (c *Config) GetInputData() []string {
	return c.Preload.InputData
}

func (c *Config) GetRequestDelay() time.Duration {
	return c.RequestDelay
}

func (c *Config) GetNetwork() string {
	if c.Network == "" {
		panic("network is not set, please set `mainnet` or `sepolia`")
	}
	return c.Network
}

func (c *Config) GetSubgraphPath() string {
	if c.SubgraphPath == "" {
		panic("SubgraphPath is not set")
	}
	return c.SubgraphPath
}

func (c *Config) GetUpstreamURL() string {
	if c.UpstreamURL == "" {
		panic("UpstreamURL is not set")
	}
	return c.UpstreamURL
}

func (c *Config) GetUpdateDelay() time.Duration {
	return c.UpdateDelay
}

func (c *Config) GetIsDebug() bool {
	return c.Debug
}

func CreateConfig() *Config {
	v := viper.New()

	viper.AutomaticEnv()

	stage := strings.TrimSpace(viper.GetString("STAGE"))
	if stage == "" {
		panic("env: STAGE is not set, please set dev,prod or local stage")
		//stage = "local"
	}

	var cfg Config
	var err error
	var data []byte

	contractsPath := strings.TrimSpace(viper.GetString("CONTRACTS"))
	if contractsPath != "" {
		data, err = os.ReadFile(contractsPath)
		if err != nil {
			panic(fmt.Errorf("incorrect CONTRACTS file: %w", err))
		}
		if err := json.Unmarshal(data, &cfg.Preload); err != nil {
			panic(fmt.Errorf("incorrect json data in CONTRACTS file: %w", err))
		}
	}

	if stage == stageLocal {
		v.AddConfigPath("../../.")
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		err := v.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}

		err = v.Unmarshal(&cfg)
		if err != nil {
			panic(fmt.Errorf("fatal error config file: %w", err))
		}

		return &cfg
	}

	return &cfg
}

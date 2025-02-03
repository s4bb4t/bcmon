package config

import (
	"encoding/json"
	"fmt"
	"git.web3gate.ru/web3/nft/GraphForge/internal/vault"
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

	Sepolia Network `mapstructure:"sepolia" json:"sepolia"`
	Holesky Network `mapstructure:"holesky" json:"holesky"`

	Debug        bool
	SubgraphPath string `mapstructure:"subgraph_path" json:"subgraph_path"`
}

type Network struct {
	Name         string `mapstructure:"name" json:"name"`
	GraphNodeURL string `mapstructure:"graph_node_url" json:"graph_node_url"`
	UpstreamURL  string `mapstructure:"upstream_url" json:"upstream_url"`

	RequestDelay time.Duration `mapstructure:"request_delay" json:"request_delay"`
	UpdateDelay  time.Duration `mapstructure:"update_delay" json:"update_delay"`
}

// Network methods

func (c *Network) ValidateNetwork() error {
	switch {
	case c.GraphNodeURL == "":
		return fmt.Errorf("incorrect graph node url")
	case c.UpstreamURL == "":
		return fmt.Errorf("incorrect upstream url")
	case c.Name == "":
		return fmt.Errorf("incorrect network name")
	default:
		return nil
	}
}

func (c *Network) GetGraphNodeURL() string {
	if c.GraphNodeURL == "" {
		panic("network is not set, please set `mainnet` or `sepolia`")
	}
	return c.GraphNodeURL
}

func (c *Network) GetRequestDelay() time.Duration {
	return c.RequestDelay
}

func (c *Network) GetNetwork() string {
	if c.Name == "" {
		panic("network name is not set, please set `mainnet` or `sepolia`")
	}
	return c.Name
}

func (c *Network) GetUpstreamURL() string {
	if c.UpstreamURL == "" {
		panic("UpstreamURL is not set")
	}
	return c.UpstreamURL
}

func (c *Network) GetUpdateDelay() time.Duration {
	return c.UpdateDelay
}

// Config methods

func (c *Config) GetSubgraphPath() string {
	if c.SubgraphPath == "" {
		panic("SubgraphPath is not set")
	}
	return c.SubgraphPath
}

func (c *Config) GetInputData() []string {
	return c.Preload.InputData
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
	debug := false
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
		debug = true
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

	if stage == stageProd {
		err := os.Setenv("LOG_LEVEL", "prod")
		if err != nil {
			panic("cannot set LOG_LEVEL env")
		}
		debug = false
	}

	secretId := viper.GetString("VAULT_SECRET_ID")
	roleId := viper.GetString("VAULT_ROLE_ID")
	vaultAddress := viper.GetString("VAULT_ADDRESS")
	vaultSecretPAth := viper.GetString("VAULT_SECRET_PATH")

	vault := vault.NewClient(vaultAddress, secretId, roleId, time.Second*5)

	secrets, err := vault.GetSecrets(vaultSecretPAth)
	if err != nil {
		panic(fmt.Errorf("cannot get vault secrets", err))
	}

	cfgBytes, err := json.Marshal(secrets)
	if err != nil {
		panic("cannot read config")
	}
	err = json.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		panic(fmt.Errorf("cannot unmarshal config: %w", err))
	}

	cfg.Debug = debug

	return &cfg
}

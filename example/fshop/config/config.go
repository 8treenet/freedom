package config

import (
	"os"
	"sync"

	"github.com/8treenet/freedom"
)

func init() {
	EntryPoint()
}

// Get .
func Get() *Configuration {
	once.Do(func() {
		cfg = newConfig()
	})
	return cfg
}

var once sync.Once
var cfg *Configuration

// Configuration .
type Configuration struct {
	App   freedom.Configuration
	DB    DBConf                 `toml:"db" yaml:"db"`
	Other map[string]interface{} `toml:"other" yaml:"other"`
	Redis RedisConf              `toml:"redis" yaml:"redis"`
}

// DBConf .
type DBConf struct {
	Addr            string `toml:"addr" yaml:"addr"`
	MaxOpenConns    int    `toml:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifeTime int    `toml:"conn_max_life_time" yaml:"conn_max_life_time"`
}

// RedisConf .
type RedisConf struct {
	Addr               string `toml:"addr" yaml:"addr"`
	Password           string `toml:"password" yaml:"password"`
	DB                 int    `toml:"db" yaml:"db"`
	MaxRetries         int    `toml:"max_retries" yaml:"max_retries"`
	PoolSize           int    `toml:"pool_size" yaml:"pool_size"`
	ReadTimeout        int    `toml:"read_timeout" yaml:"read_timeout"`
	WriteTimeout       int    `toml:"write_timeout" yaml:"write_timeout"`
	IdleTimeout        int    `toml:"idle_timeout" yaml:"idle_timeout"`
	IdleCheckFrequency int    `toml:"idle_check_frequency" yaml:"idle_check_frequency"`
	MaxConnAge         int    `toml:"max_conn_age" yaml:"max_conn_age"`
	PoolTimeout        int    `toml:"pool_timeout" yaml:"pool_timeout"`
}

func newConfig() *Configuration {
	result := &Configuration{}
	def := freedom.DefaultConfiguration()
	def.Other["listen_addr"] = ":8000"
	def.Other["service_name"] = "default"
	result.App = def

	err := freedom.Configure(&result, "config.toml")
	// err := freedom.Configure(&result, "config.yaml")
	if err == nil {
		result.App.Other = result.Other
	}
	if err != nil {
		freedom.Logger().Error(err)
	}
	return result
}

// EntryPoint .
func EntryPoint() {
	envConfig := os.Getenv("FSHOP-CONFIG")
	if envConfig != "" {
		os.Setenv(freedom.ProfileENV, envConfig)
		return
	}

	// [./main -c ./config]
	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] != "-c" {
			continue
		}
		if i+1 >= len(os.Args) {
			break
		}
		os.Setenv(freedom.ProfileENV, os.Args[i+1])
	}
}

package project

import "strings"

func init() {
	content["/server/conf/config.go"] = confTemplate()
}

func confTemplate() string {
	result := `
	package conf

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
		DB    DBConf                 $$wavetoml:"db" yaml:"db"$$wave
		Other map[string]interface{} $$wavetoml:"other" yaml:"other"$$wave
		Redis RedisConf              $$wavetoml:"redis" yaml:"redis"$$wave
	}
	
	// DBConf .
	type DBConf struct {
		Addr            string $$wavetoml:"addr" yaml:"addr"$$wave
		MaxOpenConns    int    $$wavetoml:"max_open_conns" yaml:"max_open_conns"$$wave
		MaxIdleConns    int    $$wavetoml:"max_idle_conns" yaml:"max_idle_conns"$$wave
		ConnMaxLifeTime int    $$wavetoml:"conn_max_life_time" yaml:"conn_max_life_time"$$wave
	}
	
	// RedisConf .
	type RedisConf struct {
		Addr               string $$wavetoml:"addr" yaml:"addr"$$wave
		Password           string $$wavetoml:"password" yaml:"password"$$wave
		DB                 int    $$wavetoml:"db" yaml:"db"$$wave
		MaxRetries         int    $$wavetoml:"max_retries" yaml:"max_retries"$$wave
		PoolSize           int    $$wavetoml:"pool_size" yaml:"pool_size"$$wave
		ReadTimeout        int    $$wavetoml:"read_timeout" yaml:"read_timeout"$$wave
		WriteTimeout       int    $$wavetoml:"write_timeout" yaml:"write_timeout"$$wave
		IdleTimeout        int    $$wavetoml:"idle_timeout" yaml:"idle_timeout"$$wave
		IdleCheckFrequency int    $$wavetoml:"idle_check_frequency" yaml:"idle_check_frequency"$$wave
		MaxConnAge         int    $$wavetoml:"max_conn_age" yaml:"max_conn_age"$$wave
		PoolTimeout        int    $$wavetoml:"pool_timeout" yaml:"pool_timeout"$$wave
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
		envConfig := os.Getenv("{{.PackageName}}-CONF")
		if envConfig != "" {
			os.Setenv(freedom.ProfileENV, envConfig)
			return
		}
	
		// [./base -c ./server/conf]
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
	`

	result = strings.ReplaceAll(result, "$$wave", "`")
	return result
}

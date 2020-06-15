package project

import "strings"

func init() {
	content["/server/conf/conf.go"] = confTemplate()
}

func confTemplate() string {
	result := `
	package conf

	import (
		"runtime"

		"github.com/8treenet/freedom"
	)

	func init() {
		cfg = &Configuration{
			DB:    newDBConf(),
			App:   newAppConf(),
			Redis: newRedisConf(),
		}
	}

	// Get .
	func Get() *Configuration {
		return cfg
	}

	var cfg *Configuration

	// Configuration .
	type Configuration struct {
		DB    *DBConf
		App   *freedom.Configuration
		Redis *RedisConf
	}

	// DBConf .
	type DBConf struct {
		Addr            string $$wavetoml:"addr"$$wave
		MaxOpenConns    int    $$wavetoml:"max_open_conns"$$wave
		MaxIdleConns    int    $$wavetoml:"max_idle_conns"$$wave
		ConnMaxLifeTime int    $$wavetoml:"conn_max_life_time"$$wave
	}

	// RedisConf .
	type RedisConf struct {
		Addr               string $$wavetoml:"addr"$$wave
		Password           string $$wavetoml:"password"$$wave
		DB                 int    $$wavetoml:"db"$$wave
		MaxRetries         int    $$wavetoml:"max_retries"$$wave
		PoolSize           int    $$wavetoml:"pool_size"$$wave
		ReadTimeout        int    $$wavetoml:"read_timeout"$$wave
		WriteTimeout       int    $$wavetoml:"write_timeout"$$wave
		IdleTimeout        int    $$wavetoml:"idle_timeout"$$wave
		IdleCheckFrequency int    $$wavetoml:"idle_check_frequency"$$wave
		MaxConnAge         int    $$wavetoml:"max_conn_age"$$wave
		PoolTimeout        int    $$wavetoml:"pool_timeout"$$wave
	}

	func newAppConf() *freedom.Configuration {
		result := freedom.DefaultConfiguration()
		result.Other["listen_addr"] = ":8000"
		result.Other["service_name"] = "default"
		freedom.Configure(&result, "app.toml", false)
		return &result
	}

	func newDBConf() *DBConf {
		result := &DBConf{}
		freedom.Configure(result, "db.toml", false)
		return result
	}

	func newRedisConf() *RedisConf {
		result := &RedisConf{
			MaxRetries:         0,
			PoolSize:           10 * runtime.NumCPU(),
			ReadTimeout:        3,
			WriteTimeout:       3,
			IdleTimeout:        300,
			IdleCheckFrequency: 60,
		}
		freedom.Configure(result, "redis.toml", true)
		return result
	}	`

	result = strings.ReplaceAll(result, "$$wave", "`")
	return result
}

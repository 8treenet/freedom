package project

import "fmt"

func init() {
	content["/components/config/config.go"] = confTemplate()
	content["/components/config/app.go"] = appConfTemplate()
	content["/components/config/db.go"] = dbConfTemplate()
	content["/components/config/redis.go"] = redisConfTemplate()
}

func confTemplate() string {
	return `package config

	import (
		"github.com/kataras/iris"
	)
	
	func init() {
		cfg = &Configuration{
			DB:    newDBConf(),
			App:   newAppConf(),
			Redis: newRedisConf(),
		}
	}
	
	var cfg *Configuration
	
	// Configuration .
	type Configuration struct {
		DB    *DBConf
		App   *iris.Configuration
		Redis *RedisConf
	}
	
	// Get .
	func Get() *Configuration {
		return cfg
	}
	`
}
func appConfTemplate() string {
	return `package config

	import (
		"github.com/kataras/iris"
		"github.com/8treenet/freedom"
	)
	
	func newAppConf() *iris.Configuration {
		result := iris.DefaultConfiguration()
		result.Other["listen_addr"] = ":8000"
		result.Other["service_name"] = "default"
		result.Other["trace_key"] = "Trace-ID"
		freedom.Configure(&result, "app.toml", false)
		return &result
	}
	`
}

func dbConfTemplate() string {
	result := `package config

	import "github.com/8treenet/freedom"

	func newDBConf() *DBConf {
		result := &DBConf{}
		freedom.Configure(result, "db.toml", false)
		return result
	}
	
	
	// DBCacheConf .
	type DBCacheConf struct {
		RedisConf
		Expires int %stoml:"expires"%s
	}
	
	// DBConf .
	type DBConf struct {
		Addr            string      %stoml:"addr"%s
		MaxOpenConns    int         %stoml:"max_open_conns"%s
		MaxIdleConns    int         %stoml:"max_idle_conns"%s
		ConnMaxLifeTime int         %stoml:"conn_max_life_time"%s
		Cache           DBCacheConf %stoml:"cache"%s
	}`

	list := []interface{}{}
	for index := 0; index < 12; index++ {
		list = append(list, "`")
	}
	return fmt.Sprintf(result, list...)
}

func redisConfTemplate() string {
	result := `package config

	import (
		"runtime"
		"github.com/8treenet/freedom"
	)
	
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
	}
	
	// RedisConf .
	type RedisConf struct {
		Addr               string %stoml:"addr"%s
		Password           string %stoml:"password"%s
		DB                 int    %stoml:"db"%s
		MaxRetries         int    %stoml:"max_retries"%s
		PoolSize           int    %stoml:"pool_size"%s
		ReadTimeout        int    %stoml:"read_timeout"%s
		WriteTimeout       int    %stoml:"write_timeout"%s
		IdleTimeout        int    %stoml:"idle_timeout"%s
		IdleCheckFrequency int    %stoml:"idle_check_frequency"%s
		MaxConnAge         int    %stoml:"max_conn_age"%s
		PoolTimeout        int    %stoml:"pool_timeout"%s
	}
	`
	list := []interface{}{}
	for index := 0; index < 22; index++ {
		list = append(list, "`")
	}
	return fmt.Sprintf(result, list...)
}

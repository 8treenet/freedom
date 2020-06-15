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
	Addr            string `toml:"addr"`
	MaxOpenConns    int    `toml:"max_open_conns"`
	MaxIdleConns    int    `toml:"max_idle_conns"`
	ConnMaxLifeTime int    `toml:"conn_max_life_time"`
}

// RedisConf .
type RedisConf struct {
	Addr               string `toml:"addr"`
	Password           string `toml:"password"`
	DB                 int    `toml:"db"`
	MaxRetries         int    `toml:"max_retries"`
	PoolSize           int    `toml:"pool_size"`
	ReadTimeout        int    `toml:"read_timeout"`
	WriteTimeout       int    `toml:"write_timeout"`
	IdleTimeout        int    `toml:"idle_timeout"`
	IdleCheckFrequency int    `toml:"idle_check_frequency"`
	MaxConnAge         int    `toml:"max_conn_age"`
	PoolTimeout        int    `toml:"pool_timeout"`
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
}

package config

import (
	"github.com/8treenet/freedom"
	"runtime"
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

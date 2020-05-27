package config

import (
	"github.com/8treenet/freedom"
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
	App   *freedom.Configuration
	Redis *RedisConf
}

// Get .
func Get() *Configuration {
	return cfg
}

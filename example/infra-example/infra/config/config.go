package config

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

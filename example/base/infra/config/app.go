package config

import (
	"github.com/8treenet/freedom"
)

func newAppConf() *freedom.Configuration {
	result := freedom.DefaultConfiguration()
	result.Other["listen_addr"] = ":8000"
	result.Other["service_name"] = "default"
	freedom.Configure(&result, "app.toml", false)
	return &result
}

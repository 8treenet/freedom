package config

import (
	"github.com/8treenet/freedom"
	iris "github.com/kataras/iris/v12"
)

func newAppConf() *iris.Configuration {
	result := iris.DefaultConfiguration()
	result.Other["listen_addr"] = ":8000"
	result.Other["service_name"] = "default"
	freedom.Configure(&result, "app.toml", false)
	return &result
}

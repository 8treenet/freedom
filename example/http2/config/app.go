package config

import (
	"github.com/8treenet/freedom"
	"github.com/kataras/iris"
)

func newAppConf() *iris.Configuration {
	result := iris.DefaultConfiguration()
	result.Other["listen_addr"] = ":8000"
	result.Other["service_name"] = "default"
	result.Other["trace_key"] = "Trace-ID"
	freedom.Configure(&result, "app.toml", false)
	return &result
}

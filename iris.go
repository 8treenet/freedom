package freedom

import (
	"github.com/kataras/iris/v12"
)

// DefaultConfiguration proxy a call to iris.DefaultConfiguration
func DefaultConfiguration() iris.Configuration {
	return iris.DefaultConfiguration()
}
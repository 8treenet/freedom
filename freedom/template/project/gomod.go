package project

func init() {
	content["/go.mod"] = modTemplate()
}

func modTemplate() string {
	return `
module {{.PackagePath}}

go 1.13

require (
	github.com/8treenet/freedom {{.VersionNum}}
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/kataras/iris/v12 v12.1.8
	gopkg.in/go-playground/validator.v9 v9.31.0
)

`
}

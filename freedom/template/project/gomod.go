package project

func init() {
	content["/go.mod"] = modTemplate()
}

func modTemplate() string {
	return `
module {{.PackageName}}

go 1.12

require (
	github.com/8treenet/freedom v0.0.1
	github.com/BurntSushi/toml v0.3.1
	github.com/go-redis/redis v6.15.6+incompatible // indirect
	github.com/kataras/iris v11.1.1+incompatible
	github.com/sirupsen/logrus v1.4.2 // indirect
)

`
}

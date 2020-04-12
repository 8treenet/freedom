package project

func init() {
	content["/go.mod"] = modTemplate()
}

func modTemplate() string {
	return `
module {{.PackageName}}

go 1.13

require (
	github.com/8treenet/freedom {{.VersionNum}}
	github.com/8treenet/gcache v1.1.4
	github.com/8treenet/extjson v1.0.1
	github.com/BurntSushi/toml v0.3.1
	github.com/Joker/jade v1.0.0 // indirect
	github.com/Shopify/goreferrer v0.0.0-20181106222321-ec9c9a553398 // indirect
	github.com/Shopify/sarama v1.25.0
	github.com/ajg/form v1.5.1 // indirect
	github.com/aymerick/raymond v2.0.2+incompatible // indirect
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/iris-contrib/go.uuid v2.0.0+incompatible
	github.com/jinzhu/gorm v1.9.12
	github.com/kataras/golog v0.0.10
	github.com/kataras/iris v11.1.1+incompatible
	github.com/prometheus/client_golang v1.3.0
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/sirupsen/logrus v1.4.2
)

`
}

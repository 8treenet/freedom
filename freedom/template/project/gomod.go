package project

func init() {
	content["/go.mod"] = modTemplate()
}

func modTemplate() string {
	return `
module {{.PackageName}}

go 1.13

require (
	github.com/8treenet/freedom v1.3.1
	github.com/8treenet/gcache v1.1.3
	github.com/BurntSushi/toml v0.3.1
	github.com/Joker/jade v1.0.0 // indirect
	github.com/Shopify/goreferrer v0.0.0-20181106222321-ec9c9a553398 // indirect
	github.com/Shopify/sarama v1.25.0
	github.com/ajg/form v1.5.1 // indirect
	github.com/aymerick/raymond v2.0.2+incompatible // indirect
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/flosch/pongo2 v0.0.0-20190707114632-bbf5a6c351f4 // indirect
	github.com/gavv/monotime v0.0.0-20190418164738-30dba4353424 // indirect
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/go-sql-driver/mysql v1.4.1
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/gorilla/schema v1.1.0 // indirect
	github.com/imkira/go-interpol v1.1.0 // indirect
	github.com/iris-contrib/blackfriday v2.0.0+incompatible // indirect
	github.com/iris-contrib/formBinder v5.0.0+incompatible // indirect
	github.com/iris-contrib/go.uuid v2.0.0+incompatible
	github.com/iris-contrib/httpexpect v0.0.0-20180314041918-ebe99fcebbce // indirect
	github.com/jinzhu/gorm v1.9.11
	github.com/k0kubun/colorstring v0.0.0-20150214042306-9440f1994b88 // indirect
	github.com/kataras/golog v0.0.9
	github.com/kataras/iris v11.1.1+incompatible
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/microcosm-cc/bluemonday v1.0.2 // indirect
	github.com/moul/http2curl v1.0.0 // indirect
	github.com/prometheus/client_golang v1.2.1
	github.com/ryanuber/columnize v2.1.0+incompatible
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/smartystreets/goconvey v1.6.4 // indirect
	github.com/spf13/cobra v0.0.5
	github.com/valyala/fasthttp v1.6.0
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/yudai/pp v2.0.1+incompatible // indirect
	golang.org/x/net v0.0.0-20191124235446-72fef5d5e266
)

`
}

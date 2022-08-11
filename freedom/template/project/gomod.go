package project

func init() {
	content["/go.mod"] = modTemplate()
}

func modTemplate() string {
	return `
module {{.PackagePath}}

go 1.18

require (
	github.com/8treenet/freedom {{.VersionNum}}
)

`
}

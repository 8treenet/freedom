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
)

`
}

package project

func init() {
	content["/domain/vo/vo.go"] = voTemplate()
}

func voTemplate() string {
	return `
	//Package vo generated by 'freedom new-project {{.PackagePath}}'
	package vo
	`
}

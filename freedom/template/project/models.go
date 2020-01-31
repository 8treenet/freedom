package project

func init() {
	content["/application/objects/objects.go"] = objectsTemplate()
}

func objectsTemplate() string {
	return `package objects
	`
}

package project

func init() {
	content["/application/object/object.go"] = objectsTemplate()
}

func objectsTemplate() string {
	return `package object
	`
}

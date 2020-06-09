package project

func init() {
	content["/domain/object/object.go"] = objectsTemplate()
}

func objectsTemplate() string {
	return `package object
	`
}

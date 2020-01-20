package project

func init() {
	content["/objects/objects.go"] = objectsTemplate()
}

func objectsTemplate() string {
	return `package objects
	`
}

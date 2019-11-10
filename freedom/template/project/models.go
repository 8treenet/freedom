package project

func init() {
	content["/models/models.go"] = modelsTemplate()
}

func modelsTemplate() string {
	return `package models

	import "github.com/jinzhu/gorm"
	`
}

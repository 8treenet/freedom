package project

func init() {
	content["/application/dto/dto.go"] = dtoTemplate()
}

func dtoTemplate() string {
	return `package dto
	`
}

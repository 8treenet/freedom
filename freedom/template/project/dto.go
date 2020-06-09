package project

func init() {
	content["/adapter/dto/dto.go"] = dtoTemplate()
}

func dtoTemplate() string {
	return `package dto
	`
}

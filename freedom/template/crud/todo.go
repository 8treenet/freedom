package crud

func CrudTemplate() string {

	return `
package objects
import (
	"github.com/8treenet/freedom"
	{{- if .Time}}
	"time"
	{{- end}}
)

// {{.Name}} .
{{.Content}}

// Find{{.Name}}ByPrimary .
func Find{{.Name}}ByPrimary(rep freedom.GORMRepository, primary interface{}) (result {{.Name}}, e error) {
	e = rep.DB().Find(&result, primary).Error
	return
}

// Find{{.Name}}sByPrimarys .
func Find{{.Name}}sByPrimarys(rep freedom.GORMRepository, primarys ...interface{}) (results []{{.Name}}, e error) {
	e = rep.DB().Find(&results, primarys).Error
	return
}

// Find{{.Name}} .
func Find{{.Name}}(rep freedom.GORMRepository, query {{.Name}}, builders ...freedom.QueryBuilder) (result {{.Name}}, e error) {
	db := rep.DB().Where(query)
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
	return
}

// Find{{.Name}}ByWhere .
func Find{{.Name}}ByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (result {{.Name}}, e error) {
	db := rep.DB()
	if query != "" {
		db = db.Where(query, args...)
	}
	if len(builders) == 0 {
		e = db.Last(&result).Error
		return
	}

	e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
	return
}

// Find{{.Name}}s .
func Find{{.Name}}s(rep freedom.GORMRepository, query {{.Name}}, builders ...freedom.QueryBuilder) (results []{{.Name}}, e error) {
	db := rep.DB().Where(query)

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// Find{{.Name}}sByWhere .
func Find{{.Name}}sByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (results []{{.Name}}, e error) {
	db := rep.DB()
	if query != "" {
		db = db.Where(query, args...)
	}

	if len(builders) == 0 {
		e = db.Find(&results).Error
		return
	}
	e = builders[0].Execute(db, &results)
	return
}

// Create{{.Name}} .
func Create{{.Name}}(rep freedom.GORMRepository, entity *{{.Name}}) (rowsAffected int64, e error) {
	db := rep.DB().Create(entity)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// Update{{.Name}} .
func Update{{.Name}}(rep freedom.GORMRepository, query *{{.Name}}, value {{.Name}}) (affected int64, e error) {
	db := rep.DB().Model(query).Updates(value)
	e = db.Error
	affected = db.RowsAffected
	return
}

// FindToUpdate{{.Name}}s .
func FindToUpdate{{.Name}}s(rep freedom.GORMRepository, query {{.Name}}, value {{.Name}}, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&{{.Name}}{}).Where(query).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&{{.Name}}{}).Where(query).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

// FindByWhereToUpdate{{.Name}}s .
func FindByWhereToUpdate{{.Name}}s(rep freedom.GORMRepository, query string, args []interface{}, value {{.Name}}, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&{{.Name}}{}).Where(query, args...).Order(builders[0].Order()).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&{{.Name}}{}).Where(query, args...).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}`

}

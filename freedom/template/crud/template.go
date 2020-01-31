package crud

func CrudTemplate() string {

	return `
package objects
{{- if .Time}}
import (
	"time"
)
{{- end}}

// {{.Name}} .
{{.Content}}

// Save .
func (obj *{{.Name}})Save() map[string]interface{} {
	if obj.changes == nil {
		return nil
	}
	result := make(map[string]interface{})
	for k, v := range obj.changes {
		result[k] = v
	}
	obj.changes = nil
	return result
}

{{range .Fields}}
// Set{{.Value}} .
func (obj *{{.StructName}}) Set{{.Value}} ({{.Arg}} {{.Type}}) {
	if obj.changes == nil {
		obj.changes = make(map[string]interface{})
	}
	obj.{{.Value}} = {{.Arg}} 
	obj.changes["{{.Column}}"] = {{.Arg}}
}
{{ end }}
`
}

func FunTemplatePackage() string {
	return `
	package repositorys
	import (
		"github.com/8treenet/freedom"
		"{{.PackagePath}}"
	)
`
}
func FunTemplate() string {
	return `

	// find{{.Name}}ByPrimary .
	func find{{.Name}}ByPrimary(rep freedom.GORMRepository, primary interface{}) (result objects.{{.Name}}, e error) {
		e = rep.DB().Find(&result, primary).Error
		return
	}
	
	// find{{.Name}}sByPrimarys .
	func find{{.Name}}sByPrimarys(rep freedom.GORMRepository, primarys ...interface{}) (results []objects.{{.Name}}, e error) {
		e = rep.DB().Find(&results, primarys).Error
		return
	}
	
	// find{{.Name}} .
	func find{{.Name}}(rep freedom.GORMRepository, query objects.{{.Name}}, builders ...freedom.QueryBuilder) (result objects.{{.Name}}, e error) {
		db := rep.DB().Where(query)
		if len(builders) == 0 {
			e = db.Last(&result).Error
			return
		}
	
		e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
		return
	}
	
	// find{{.Name}}ByWhere .
	func find{{.Name}}ByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (result objects.{{.Name}}, e error) {
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
	
	// find{{.Name}}ByMap .
	func find{{.Name}}ByMap(rep freedom.GORMRepository, query map[string]interface{}, builders ...freedom.QueryBuilder) (result objects.{{.Name}}, e error) {
		db := rep.DB().Where(query)
		if len(builders) == 0 {
			e = db.Last(&result).Error
			return
		}
	
		e = db.Limit(1).Order(builders[0].Order()).Find(&result).Error
		return
	}
	
	// find{{.Name}}s .
	func find{{.Name}}s(rep freedom.GORMRepository, query objects.{{.Name}}, builders ...freedom.QueryBuilder) (results []objects.{{.Name}}, e error) {
		db := rep.DB().Where(query)
	
		if len(builders) == 0 {
			e = db.Find(&results).Error
			return
		}
		e = builders[0].Execute(db, &results)
		return
	}
	
	// find{{.Name}}sByWhere .
	func find{{.Name}}sByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (results []objects.{{.Name}}, e error) {
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
	
	// find{{.Name}}sByMap .
	func find{{.Name}}sByMap(rep freedom.GORMRepository, query map[string]interface{}, builders ...freedom.QueryBuilder) (results []objects.{{.Name}}, e error) {
		db := rep.DB().Where(query)
	
		if len(builders) == 0 {
			e = db.Find(&results).Error
			return
		}
		e = builders[0].Execute(db, &results)
		return
	}
	
	// create{{.Name}} .
	func create{{.Name}}(rep freedom.GORMRepository, object *objects.{{.Name}}) (rowsAffected int64, e error) {
		db := rep.DB().Create(object)
		rowsAffected = db.RowsAffected
		e = db.Error
		return
	}

	// update{{.Name}} .
	func update{{.Name}}(rep freedom.GORMRepository, object *objects.{{.Name}}) (affected int64, e error) {
		db := rep.DB().Model(object).Updates(object.Save())
		e = db.Error
		affected = db.RowsAffected
		return
	}
`
}

package models

import (
	"github.com/8treenet/freedom"
)

// Todo .
type Todo struct {
}

// FindTodoByPrimary .
func FindTodoByPrimary(rep freedom.GORMRepository, primary interface{}) (result Todo, e error) {
	e = rep.DB().Find(&result, primary).Error
	return
}

// FindTodosByPrimarys .
func FindTodosByPrimarys(rep freedom.GORMRepository, primarys []interface{}) (results []Todo, e error) {
	e = rep.DB().Find(results, primarys).Error
	return
}

// FindTodo .
func FindTodo(rep freedom.GORMRepository, query *Todo, builders ...freedom.QueryBuilder) (result Todo, e error) {
	if len(builders) == 0 {
		e = rep.DB().Where(query).Last(&result).Error
		return
	}

	e = rep.DB().Where(query).Limit(1).Order(builders[0].Order()).Find(&result).Error
	return
}

// FindTodos .
func FindTodos(rep freedom.GORMRepository, query *Todo, builders ...freedom.QueryBuilder) (results []Todo, e error) {
	db := rep.DB()
	if len(builders) == 0 {
		e = db.Where(query).Find(&results).Error
		return
	}

	where := db.Where(query)
	e = builders[0].Execute(where, &results)
	return
}

// FindTodosByWhere .
func FindTodosByWhere(rep freedom.GORMRepository, query string, args []interface{}, builders ...freedom.QueryBuilder) (results []Todo, e error) {
	db := rep.DB()
	if len(builders) == 0 {
		e = db.Where(query, args...).Find(&results).Error
		return
	}

	where := db.Where(query, args...)
	e = builders[0].Execute(where, &results)
	return
}

// CreateTodo .
func CreateTodo(rep freedom.GORMRepository, entity *Todo) (rowsAffected int64, e error) {
	db := rep.DB().Create(entity)
	rowsAffected = db.RowsAffected
	e = db.Error
	return
}

// UpdateTodo .
func UpdateTodo(rep freedom.GORMRepository, entity *Todo, value Todo) (affected int64, e error) {
	db := rep.DB().Model(entity).Updates(value)
	e = db.Error
	affected = db.RowsAffected
	return
}

// FindToUpdateTodos .
func FindToUpdateTodos(rep freedom.GORMRepository, query *Todo, value Todo, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Todo{}).Where(query).Order(builders[0].Order).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&Todo{}).Where(query).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

// FindByWhereToUpdateTodos .
func FindByWhereToUpdateTodos(rep freedom.GORMRepository, query string, args []interface{}, value Todo, builders ...freedom.QueryBuilder) (affected int64, e error) {
	db := rep.DB()
	if len(builders) > 0 {
		db = db.Model(&Todo{}).Where(query, args...).Order(builders[0].Order).Limit(builders[0].Limit()).Updates(value)
	} else {
		db = db.Model(&Todo{}).Where(query, args...).Updates(value)
	}

	affected = db.RowsAffected
	e = db.Error
	return
}

package entity

import (
	"strconv"

	"github.com/8treenet/freedom/example/fshop/domain/object"

	"github.com/8treenet/freedom"
)

// 管理员实体
type Admin struct {
	freedom.Entity
	object.Admin
}

// Identity 唯一
func (admin *Admin) Identity() string {
	return strconv.Itoa(admin.Id)
}

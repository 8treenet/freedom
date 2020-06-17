package entity

import (
	"strconv"

	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/fshop/adapter/po"
)

// 管理员实体
type Admin struct {
	freedom.Entity
	po.Admin
}

// Identity 唯一
func (admin *Admin) Identity() string {
	return strconv.Itoa(admin.Id)
}

package entity

import (
	"errors"
	"strconv"

	"github.com/8treenet/freedom/example/fshop/domain/object"

	"github.com/8treenet/freedom"
)

// 用户实体
type User struct {
	freedom.Entity
	object.User
}

// Identity 唯一
func (u *User) Identity() string {
	return strconv.Itoa(u.Id)
}

// ChangePassword 修改密码
func (u *User) ChangePassword(newPassword, oldPassword string) error {
	if u.Password != oldPassword {
		return errors.New("Password error")
	}
	u.SetPassword(newPassword)
	return nil
}

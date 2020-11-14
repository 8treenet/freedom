package event

import "encoding/json"

// ChangePassword 修改密码事件
type ChangePassword struct {
	ID          int `json:"id"`
	prototypes  map[string]interface{}
	UserID      int    `json:"userID"`
	NewPassword string `json:"newPassword"`
	OldPassword string `json:"oldPassword"`
}

// Topic .
func (password *ChangePassword) Topic() string {
	return "ChangePassword"
}

// SetPrototypes .
func (password *ChangePassword) SetPrototypes(prototypes map[string]interface{}) {
	if password.prototypes == nil {
		password.prototypes = prototypes
		return
	}

	for key, value := range prototypes {
		password.prototypes[key] = value
	}
}

// GetPrototypes .
func (password *ChangePassword) GetPrototypes() map[string]interface{} {
	return password.prototypes
}

// Marshal .
func (password *ChangePassword) Marshal() []byte {
	data, _ := json.Marshal(password)
	return data
}

// Identity .
func (password *ChangePassword) Identity() interface{} {
	return password.ID
}

// SetIdentity .
func (password *ChangePassword) SetIdentity(identity interface{}) {
	password.ID = identity.(int)
}

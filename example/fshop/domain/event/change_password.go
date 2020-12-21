package event

import "encoding/json"

// ChangePassword 修改密码事件
type ChangePassword struct {
	ID          string `json:"identity"`
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
func (shop *ChangePassword) Marshal() ([]byte, error) {
	return json.Marshal(shop)
}

// Unmarshal .
func (shop *ChangePassword) Unmarshal(data []byte) error {
	return json.Unmarshal(data, shop)
}

// Identity .
func (password *ChangePassword) Identity() string {
	return password.ID
}

// SetIdentity .
func (password *ChangePassword) SetIdentity(identity string) {
	password.ID = identity
}

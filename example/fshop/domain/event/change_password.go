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
func (event *ChangePassword) Topic() string {
	return "ChangePassword"
}

// SetPrototypes .
func (event *ChangePassword) SetPrototypes(prototypes map[string]interface{}) {
	if event.prototypes == nil {
		event.prototypes = prototypes
		return
	}

	for key, value := range prototypes {
		event.prototypes[key] = value
	}
}

// GetPrototypes .
func (event *ChangePassword) GetPrototypes() map[string]interface{} {
	return event.prototypes
}

// Marshal .
func (event *ChangePassword) Marshal() ([]byte, error) {
	return json.Marshal(event)
}

// Unmarshal .
func (event *ChangePassword) Unmarshal(data []byte) error {
	return json.Unmarshal(data, event)
}

// Identity .
func (event *ChangePassword) Identity() string {
	return event.ID
}

// SetIdentity .
func (event *ChangePassword) SetIdentity(identity string) {
	event.ID = identity
}

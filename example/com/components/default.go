package components

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindComponent(false, func() *DefaultComponent {
			return &DefaultComponent{}
		})
	})
}

// DefaultComponent .
type DefaultComponent struct {
	freedom.Component
	Value string
}

// BeginRequest .
func (c *DefaultComponent) BeginRequest(rt freedom.Runtime) {
	c.Component.BeginRequest(rt)
	c.Value = ""

	//如果是组件要总线传递,先获取是否有该组件
	m, ok := c.GetBus().Get("DefaultComponent")
	if ok {
		c.Value = m.(map[string]interface{})["Value"].(string)
	}
	c.GetBus().Add("DefaultComponent", c)
}

// SetValue .
func (c *DefaultComponent) SetValue(v string) {
	c.Value = v
}

// GetValue .
func (c *DefaultComponent) GetValue() string {
	return c.Value
}

// Print .
func (c *DefaultComponent) Print() {
	c.Runtime.Logger().Info("DefaultComponent", c.Value)
}

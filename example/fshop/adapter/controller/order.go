package controller

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindController("/order", &Order{})
	})
}

type Order struct {
	Runtime freedom.Runtime
}

package controller

import (
	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindController("/delivery", &Delivery{})
	})
}

type Delivery struct {
	Runtime freedom.Runtime
}

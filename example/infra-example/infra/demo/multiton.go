package demo

import (
	"github.com/8treenet/freedom"
	"math/rand"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindInfra(false, func() *Multiton {
			return &Multiton{}
		})
		initiator.InjectController(func(ctx freedom.Context) (com *Multiton) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// Multiton .
type Multiton struct {
	freedom.Infra
	life int //生命
}

// BeginRequest
func (mu *Multiton) BeginRequest(rt freedom.Runtime) {
	mu.Infra.BeginRequest(rt)
	rt.Logger().Info("Multiton 初始化拉")
	mu.life = rand.Intn(100)
}

func (mu *Multiton) GetLife() int {
	return mu.life
}

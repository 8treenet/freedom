package demo

import (
	"math/rand"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, &Single{})
		initiator.InjectController(func(ctx freedom.Context) (com *Single) {
			initiator.GetInfra(ctx, &com)
			return
		})
	})
}

// Single .
type Single struct {
	life int //生命
}

// Booting .
func (c *Single) Booting(boot freedom.SingleBoot) {
	freedom.Logger().Info("Single.Booting")
	c.life = rand.Intn(100)
}

func (mu *Single) GetLife() int {
	return mu.life
}

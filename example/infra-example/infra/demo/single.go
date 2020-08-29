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
func (s *Single) Booting(boot freedom.SingleBoot) {
	freedom.Logger().Info("Single.Booting")
	s.life = rand.Intn(100)
}

// GetLife .
func (s *Single) GetLife() int {
	return s.life
}

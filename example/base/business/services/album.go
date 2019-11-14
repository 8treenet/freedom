package services

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/business/repositorys"
	"github.com/8treenet/freedom/example/base/models"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindService(func() *AlbumService {
			return &AlbumService{}
		})
	})
}

// AlbumRepoInterface .
type AlbumRepoInterface interface {
	GetUA() string
}

// AlbumService .
type AlbumService struct {
	Runtime   freedom.Runtime
	DefRepo   *repositorys.AlbumRepository
	DefRepoIF AlbumRepoInterface
}

// GetAlbum .
func (s *AlbumService) GetAlbum(id int) (*models.Album, error) {
	return s.DefRepo.GetAlbum(id)
}

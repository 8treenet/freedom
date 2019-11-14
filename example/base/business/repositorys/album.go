package repositorys

import (
	"github.com/8treenet/freedom"
	"github.com/8treenet/freedom/example/base/models"
)

func init() {
	freedom.Booting(func(initiator freedom.Initiator) {
		initiator.BindRepository(func() *AlbumRepository {
			println("initiator.BindRepository AlbumRepository")
			return &AlbumRepository{}
		})
	})
}

// AlbumRepository .
type AlbumRepository struct {
	freedom.Repository
	models.Album
}

// GetAlbum .
func (repo *AlbumRepository) GetAlbum(id int) (*models.Album, error) {
	//repo.DB().Find()
	a := &models.Album{}
	repo.Runtime.Logger().Infof("%#v\n", repo)
	if err := repo.DB().First(a, id).Error; err != nil {
		repo.Runtime.Logger().Errorf("repo.GetAlbum id: %d, error: %s", id, err.Error())
		return nil, err
	}
	repo.Runtime.Logger().Infof("repo.GetAlbum id: %d, alblum: %v", id, a)
	return a, nil
}

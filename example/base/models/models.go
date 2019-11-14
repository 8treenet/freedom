package models

// Album test
type Album struct {
	AlbumID  int    `gorm:"column:AlbumId;primary_key" json:"id"`
	Title    string `gorm:"column:Title" json:"title"`
	ArtistID int    `gorm:"column:ArtistId" json:"artistId"`
}

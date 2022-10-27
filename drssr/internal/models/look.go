package models

import "time"

//easyjson:json
type Look struct {
	ID          uint64    `json:"id" db:"id"`
	Description string    `json:"description" db:"description"`
	CreatorID   uint64    `json:"creator_id" db:"creator_id"`
	Clothes     []uint64  `json:"clothes" db:"-"`
	ImgPath     string    `json:"-" db:"img"`
	Img         string    `json:"img" db:"-"`
	PreviewPath string    `json:"-" db:"preview"`
	Preview     string    `json:"preview" db:"-"`
	Ctime       time.Time `json:"-" db:"created_at"`
}

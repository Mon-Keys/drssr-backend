package models

import "time"

//easyjson:json
type Look struct {
	ID          uint64          `json:"id" db:"id"`
	Description string          `json:"description" db:"description"`
	Filename    string          `json:"filename,omitempty"`
	CreatorID   uint64          `json:"creator_id" db:"creator_id"`
	Clothes     []ClothesStruct `json:"clothes" db:"-"`
	ImgPath     string          `json:"-" db:"img"`
	Img         string          `json:"img" db:"-"`
	PreviewPath string          `json:"-" db:"preview"`
	Preview     string          `json:"preview" db:"-"`
	Ctime       time.Time       `json:"-" db:"created_at"`
}

//easyjson:json
type ClothesStruct struct {
	ID     uint64       `json:"id" db:"id"`
	Label  string       `json:"label" db:"name"`
	Coords CoordsStruct `json:"coords" db:"coords"`
}

//easyjson:json
type CoordsStruct struct {
	X int `json:"x" db:"x"`
	Y int `json:"y" db:"y"`
}

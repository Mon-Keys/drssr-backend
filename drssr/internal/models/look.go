package models

import "time"

//easyjson:json
type Look struct {
	ID          uint64          `json:"id" db:"id"`
	Description string          `json:"description" db:"description"`
	Filename    string          `json:"filename,omitempty"`
	CreatorID   uint64          `json:"creator_id" db:"creator_id"`
	Clothes     []ClothesStruct `json:"clothes" db:"-"`
	ImgPath     string          `json:"img_path" db:"img"`
	Img         string          `json:"img" db:"-"`
	Ctime       time.Time       `json:"-" db:"created_at"`
}

//easyjson:json
type ClothesStruct struct {
	ID       uint64       `json:"id" db:"id"`
	Label    string       `json:"label" db:"name"`
	Coords   CoordsStruct `json:"coords" db:"coords"`
	ImgPath  string       `json:"img_path" db:"img_path"`
	MaskPath string       `json:"mask_path" db:"mask_path"`
}

//easyjson:json
type CoordsStruct struct {
	X int `json:"x" db:"x"`
	Y int `json:"y" db:"y"`
}

//easyjson:json
type ArrayLooks []Look

package models

import "time"

//easyjson:json
type Look struct {
	ID        uint64          `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Desc      string          `json:"description" db:"description"`
	CreatorID uint64          `json:"creator_id" db:"creator_id"`
	Clothes   []ClothesStruct `json:"clothes" db:"-"`
	ImgPath   string          `json:"img_path" db:"img"`
	Img       string          `json:"img" db:"-"`
	Ctime     time.Time       `json:"-" db:"created_at"`
}

//easyjson:json
type ClothesStruct struct {
	ID       uint64       `json:"id" db:"id"`
	Name     string       `json:"name" db:"name"`
	Desc     string       `json:"description" db:"description"`
	Type     string       `json:"type" db:"type"`
	Brand    string       `json:"brand" db:"brand"`
	Coords   CoordsStruct `json:"coords" db:"coords"`
	Rotation int          `json:"rotation"`
	Scaling  int          `json:"scaling"`
	ImgPath  string       `json:"img_path" db:"img_path"`
	MaskPath string       `json:"mask_path" db:"mask_path"`
}

//easyjson:json
type CoordsStruct struct {
	X int `json:"x" db:"x"`
	Y int `json:"y" db:"y"`
	Z int `json:"z" db:"z"`
}

//easyjson:json
type ArrayLooks []Look

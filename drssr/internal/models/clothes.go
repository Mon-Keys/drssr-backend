package models

import "time"

//easyjson:json
type Clothes struct {
	ID       uint64    `json:"id" db:"id"`
	Type     string    `json:"type" db:"type"`
	Color    string    `json:"color" db:"color"`
	ImgPath  string    `json:"-" db:"img"`
	Img      string    `json:"img" db:"-"`
	MaskPath string    `json:"-" db:"mask"`
	Mask     string    `json:"mask" db:"-"`
	Brand    string    `json:"brand" db:"brand"`
	Sex      string    `json:"sex" db:"sex"`
	Ctime    time.Time `json:"-" db:"created_at"`
}

//easyjson:json
type ArrayClothes []Clothes

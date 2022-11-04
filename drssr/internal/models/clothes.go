package models

import "time"

type CurrencyType string

const (
	CurrencyRUB CurrencyType = "RUB"
)

//easyjson:json
type Clothes struct {
	ID       uint64       `json:"id" db:"id"`
	Type     string       `json:"type" db:"type"`
	Color    string       `json:"color,omitempty" db:"color"`
	ImgPath  string       `json:"-" db:"img"`
	Img      string       `json:"img" db:"-"`
	MaskPath string       `json:"-" db:"mask"`
	Mask     string       `json:"mask" db:"-"`
	Brand    string       `json:"brand,omitempty" db:"brand"`
	Sex      string       `json:"sex,omitempty" db:"sex"`
	Ctime    time.Time    `json:"-" db:"created_at"`
	Link     string       `json:"link,omitempty" db:"link"`
	Price    uint         `json:"price,omitempty" db:"price"`
	Currency CurrencyType `json:"currency,omitempty" db:"currency"`
	OwnerID  uint64       `json:"owner_id" db:"owner_id"`
}

//easyjson:json
type ArrayClothes []Clothes

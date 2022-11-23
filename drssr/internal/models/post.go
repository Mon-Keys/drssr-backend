package models

import "time"

type PostType string

const (
	PostTypeClothes PostType = "clothes"
	PostTypeLook    PostType = "look"
)

//easyjson:json
type Post struct {
	ID            uint64    `json:"id" db:"id"`
	Type          PostType  `json:"type" db:"type"`
	Name          string    `json:"name" db:"name"`
	Desc          string    `json:"description" db:"description"`
	CreatorID     uint64    `json:"creator_id" db:"creator_id"`
	ElementID     uint64    `json:"element_id" db:"element_id"`
	Clothes       Clothes   `json:"clothes,omitempty" db:"-"`
	Look          Look      `json:"look,omitempty" db:"-"`
	Previews      []string  `json:"previews" db:"-"`
	PreviewsPaths []string  `json:"previews_paths" db:"previews_paths"`
	Likes         int       `json:"likes" db:"likes"`
	Ctime         time.Time `json:"-" db:"created_at"`
}

//easyjson:json
type ArrayPosts []Post

//easyjson:json
type LikesStruct struct {
	Likes int `json:"likes" db:"-"`
}

package models

import "time"

//easyjson:json
type User struct {
	ID       uint64    `json:"id" db:"id"`
	Nickname string    `json:"nickname" db:"nickname"`
	Email    string    `json:"email" db:"email"`
	Password string    `json:"-" db:"password"`
	Avatar   string    `json:"avatar" db:"avatar"`
	Name     string    `json:"name,omitempty" db:"name"`
	Stylist  bool      `json:"stylist" db:"stylist"`
	Age      int       `json:"age" db:"age"`
	Desc     string    `json:"description,omitempty" db:"description"`
	Ctime    time.Time `json:"-" db:"created_at"`
}

//easyjson:json
type UpdateUserReq struct {
	Nickname  string `json:"nickname" db:"nickname"`
	Email     string `json:"email" db:"email"`
	Avatar    string `json:"avatar" db:"avatar"`
	Name      string `json:"name,omitempty" db:"name"`
	Stylist   bool   `json:"-" db:"stylist"`
	BirthDate string `json:"birth_date" db:"birth_date"`
	Desc      string `json:"description,omitempty" db:"description"`
}

//easyjson:json
type SignupCredentials struct {
	Nickname  string `json:"nickname" db:"nickname"`
	Email     string `json:"email" db:"email"`
	Password  string `json:"password" db:"password"`
	Name      string `json:"name,omitempty" db:"name"`
	Avatar    string `json:"avatar" db:"avatar"`
	BirthDate string `json:"birth_date" db:"birth_date"`
	Desc      string `json:"description,omitempty" db:"description"`
}

//easyjson:json
type LoginCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

//easyjson:json
type StatusCheckStruct struct {
	UserTotal int `json:"user_total"`
}

//easyjson:json
type StylistRequest struct {
	ID    uint64    `json:"id" db:"id"`
	UID   uint64    `json:"uid" db:"user_id"`
	Ctime time.Time `json:"-" db:"created_at"`
}

//easyjson:json
type StylistRequestStatusStruct struct {
	Exists bool `json:"exists" db:"-"`
}

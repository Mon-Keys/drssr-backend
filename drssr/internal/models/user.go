package models

//easyjson:json
type User struct {
	ID        uint64 `json:"id" db:"id"`
	Nickname  string `json:"nickname" db:"nickname"`
	Email     string `json:"email" db:"email"`
	Password  string `json:"-" db:"password"`
	Avatar    string `json:"avatar" db:"avatar"`
	Name      string `json:"name,omitempty" db:"name"`
	Stylist   bool   `json:"stylist" db:"stylist"`
	BirthDate string `json:"-" db:"birth_date"`
	Desc      string `json:"description,omitempty" db:"description"`
	Ctime     uint64 `json:"-" db:"created_at"`
}

//easyjson:json
type SignupCredentials struct {
	Nickname  string `json:"nickname" db:"nickname"`
	Email     string `json:"email" db:"email"`
	Password  string `json:"-" db:"password"`
	Name      string `json:"name,omitempty" db:"name"`
	BirthDate string `json:"-" db:"birth_date"`
	Desc      string `json:"description,omitempty" db:"description"`
}

//easyjson:json
type LoginCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

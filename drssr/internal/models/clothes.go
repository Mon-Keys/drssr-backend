package models

import "time"

type CurrencyType string

const (
	CurrencyRUB CurrencyType = "RUB"
)

var (
	ToRussianType = map[string]string{
		"Anorak":      "Анорак",
		"Blazer":      "Блейзер",
		"Blouse":      "Блузка",
		"Bomber":      "Бомбер",
		"Button-Down": "На пуговицах",
		"Caftan":      "Кафтан",
		"Capris":      "Капри",
		"Cardigan":    "Кардиган",
		"Chinos":      "Чиносы",
		"Coat":        "Пальто",
		"Coverup":     "Туника",
		"Culottes":    "Брюки-кюлоты",
		"Cutoffs":     "Отсечки",
		"Dress":       "Платье",
		"Flannel":     "Фланель",
		"Gauchos":     "Гаучо",
		"Halter":      "Поводок",
		"Henley":      "Хенли",
		"Hoodie":      "Худи",
		"Jacket":      "Пиджак",
		"Jeans":       "Джинсы",
		"Jeggings":    "Джеггинсы",
		"Jersey":      "Джерси",
		"Jodhpurs":    "Джодхпуры",
		"Joggers":     "Джоггеры",
		"Jumpsuit":    "Комбинезон",
		"Kaftan":      "Кафтан",
		"Kimono":      "Кимоно",
		"Leggings":    "Леггинсы",
		"Onesie":      "Кигуруми",
		"Parka":       "Парка",
		"Peacoat":     "Бушлат",
		"Poncho":      "Пончо",
		"Robe":        "Халат",
		"Romper":      "Ромпер",
		"Sarong":      "Саронг",
		"Shorts":      "Шорты",
		"Skirt":       "Юбка",
		"Sweater":     "Свитер",
		"Sweatpants":  "Тренировочные брюки",
		"Sweatshorts": "Спортивные шорты",
		"Tank":        "Танк",
		"Tee":         "Футболка",
		"Top":         "Топ",
		"Trunks":      "Транки",
		"Turtleneck":  "Водолазка",
	}

	ToEnglishType = map[string]string{
		"Anorak":       "Anorak",
		"Blazer":       "Blazer",
		"Blouse":       "Blouse",
		"Bomber":       "Bomber",
		"Button-Down":  "Button-Down",
		"Caftan":       "Caftan",
		"Capris":       "Capris",
		"Cardigan":     "Cardigan",
		"Chinos":       "Chinos",
		"Coat":         "Coat",
		"Coverup":      "Coverup",
		"Culottes":     "Culottes",
		"Cutoffs":      "Cutoffs",
		"Dress":        "Dress",
		"Flannel":      "Flannel",
		"Gauchos":      "Gauchos",
		"Halter":       "Halter",
		"Henley":       "Henley",
		"Hoodie":       "Hoodie",
		"Jacket":       "Jacket",
		"Jeans":        "Jeans",
		"Jeggings":     "Jeggings",
		"Jersey":       "Jersey",
		"Jodhpurs":     "Jodhpurs",
		"Joggers":      "Joggers",
		"Jumpsuit":     "Jumpsuit",
		"Kaftan":       "Kaftan",
		"Kimono":       "Kimono",
		"Leggings":     "Leggings",
		"Onesie":       "Onesie",
		"Parka":        "Parka",
		"Peacoat":      "Peacoat",
		"Poncho":       "Poncho",
		"Robe":         "Robe",
		"Romper":       "Romper",
		"Sarong":       "Sarong",
		"Shorts":       "Shorts",
		"Skirt":        "Skirt",
		"Sweater":      "Sweater",
		"Sweatpants":   "Sweatpants",
		"Sweatshorts":  "Sweatshorts",
		"Tank":         "Tank",
		"Tee":          "Tee",
		"Top":          "Top",
		"Trunks":       "Trunks",
		"Turtleneck":   "Turtleneck",
		"Анорак":       "Anorak",
		"Блейзер":      "Blazer",
		"Блузка":       "Blouse",
		"Бомбер":       "Bomber",
		"На пуговицах": "Button-Down",
		"Кафтан":       "Caftan",
		"Капри":        "Capris",
		"Кардиган":     "Cardigan",
		"Чиносы":       "Chinos",
		"Пальто":       "Coat",
		"Туника":       "Coverup",
		"Брюки-кюлоты": "Culottes",
		"Отсечки":      "Cutoffs",
		"Платье":       "Dress",
		"Фланель":      "Flannel",
		"Гаучо":        "Gauchos",
		"Поводок":      "Halter",
		"Хенли":        "Henley",
		"Худи":         "Hoodie",
		"Пиджак":       "Jacket",
		"Джинсы":       "Jeans",
		"Джеггинсы":    "Jeggings",
		"Джерси":       "Jersey",
		"Джодхпуры":    "Jodhpurs",
		"Джоггеры":     "Joggers",
		"Комбинезон":   "Jumpsuit",
		// "Кафтан":              "Kaftan",
		"Кимоно":              "Kimono",
		"Леггинсы":            "Leggings",
		"Кигуруми":            "Onesie",
		"Парка":               "Parka",
		"Бушлат":              "Peacoat",
		"Пончо":               "Poncho",
		"Халат":               "Robe",
		"Ромпер":              "Romper",
		"Саронг":              "Sarong",
		"Шорты":               "Shorts",
		"Юбка":                "Skirt",
		"Свитер":              "Sweater",
		"Тренировочные брюки": "Sweatpants",
		"Спортивные шорты":    "Sweatshorts",
		"Танк":                "Tank",
		"Футболка":            "Tee",
		"Топ":                 "Top",
		"Транки":              "Trunks",
		"Водолазка":           "Turtleneck",
	}
)

//easyjson:json
type Clothes struct {
	ID       uint64       `json:"id" db:"id"`
	Name     string       `json:"name" db:"name"`
	Desc     string       `json:"description" db:"description"`
	Type     string       `json:"type" db:"type"`
	Color    string       `json:"color,omitempty" db:"color"`
	ImgPath  string       `json:"img_path" db:"img"`
	Img      string       `json:"img" db:"-"`
	MaskPath string       `json:"mask_path" db:"mask"`
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

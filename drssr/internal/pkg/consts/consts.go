package consts

// files
const (
	MaxUploadFileSize     = 20 << 20 // 5MB
	AvatarsBaseFolderPath = "./media/avatars"
	ClothesBaseFolderPath = "./media/clothes"
	MasksBaseFolderPath   = "./media/masks"
	LooksBaseFolderPath   = "./media/looks"
	PostsBaseFolderPath   = "./media/posts"
	FileExt               = "webp"

	DefaultsBaseFolderPath = "./media/defaults"
	DefaultAvatarFileName  = "avatar.webp"
)

// clothes
const (
	GetClothesLimit = 1000
)

func ClothesTypeToRussian(engType string) string {
	var russianTypes = map[string]string{
		"Anorak":      "Анорак",
		"Blazer":      "Блейзер",
		"Blouse":      "Блузка",
		"Bomber":      "Бомбер",
		"Button-Down": "Рубашка",
		"Caftan":      "Кафтан",
		"Capris":      "Капри",
		"Cardigan":    "Кардиган",
		"Chinos":      "Чиносы",
		"Coat":        "Пальто",
		"Coverup":     "Туника",
		"Culottes":    "Брюки-кюлоты",
		"Cutoffs":     "Обрезанные шорты",
		"Dress":       "Платье",
		"Flannel":     "Фланель",
		// ???
		"Gauchos":  "Гаучо",
		"Halter":   "Топ с лямкой на шее",
		"Henley":   "Пуловер с воротником",
		"Hoodie":   "Худи",
		"Jacket":   "Пиджак",
		"Jeans":    "Джинсы",
		"Jeggings": "Джеггинсы",
		"Jersey":   "Джерси",
		// вообще это шорты для езды на лошади
		"Jodhpurs":    "Бриджи",
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
		"Tank":        "Майка",
		"Tee":         "Футболка",
		"Top":         "Топ",
		"Trunks":      "Плавки",
		"Turtleneck":  "Водолазка",
	}

	rusType, ok := russianTypes[engType]
	if !ok {
		// for custom types
		return engType
	}

	return rusType
}

// looks
const (
	GetLooksLimit = 1000
)

// posts
const (
	GetPostsLimit = 1000
)

// tgBot
const (
	StylistRequestMsg        = "Запрос на подтверждения статуса стилиста #%d.\nПочта: %s\nНик: %s\nИмя: %s\nВозраст: %d\nОписание: %s"
	StylistStatusAcceptOKMsg = "Статус стилиста успешно подтвержден для пользователя %s"
	StylistStatusRejectOKMsg = "Статус стилиста отклонен для пользователя %s"
	TGBotResponseAccept      = iota
	TGBotResponseReject
)

// mailer
const (
	StylistStatusAcceptMsg = "Здравствуйте, %s. <br/> Поздравляем, ваш статус стилиста подтвержден! <br/> Теперь вы можете публиковать свои образы. <br/> С уважением, команда Kiroo."
	StylistStatusRejectMsg = "Здравствуйте, %s. <br/> К сожалению, вам отклонили заявку на статус стилиста. <br/> С уважением, команда Kiroo."
)

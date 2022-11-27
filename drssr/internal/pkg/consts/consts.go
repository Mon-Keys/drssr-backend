package consts

// files
const (
	MaxUploadFileSize     = 5 << 20 // 5MB
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
	StylistStatusAcceptMsg = "Здравствуйте, %s. <br/> Поздравляем, ваш статус стилиста подтвержден! <br/> Теперь вы можете публиковать свои образы. <br/> С уважением, команда Pose."
	StylistStatusRejectMsg = "Здравствуйте, %s. <br/> К сожалению, вам отклонили заявку на статус стилиста. <br/> С уважением, команда Pose."
)

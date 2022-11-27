package usecase

import (
	"context"
	"crypto/sha1"
	"drssr/config"
	"drssr/internal/models"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/file_utils"
	"drssr/internal/pkg/hasher"
	"drssr/internal/pkg/rollback"
	"encoding/hex"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"

	"drssr/internal/users/repository"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type IUserUsecase interface {
	GetUserByCookie(ctx context.Context, cookie string) (models.User, int, error)
	GetUserByNickname(ctx context.Context, nickname string) (models.User, int, error)
	SignupUser(ctx context.Context, credentials models.SignupCredentials) (models.User, string, int, error)
	LoginUser(ctx context.Context, credentials models.LoginCredentials) (models.User, string, int, error)
	LogoutUser(ctx context.Context, email string) (int, error)
	UpdateUser(ctx context.Context, newUserData models.UpdateUserReq) (models.User, int, error)
	DeleteUser(ctx context.Context, user models.User, cookieValue string) (int, error)
	UpdateAvatar(ctx context.Context, args UpdateAvatarArgs) (models.User, int, error)
	DeleteAvatar(ctx context.Context, user models.User) (models.User, int, error)

	CheckStatus(ctx context.Context) (int, error)

	BecomeStylist(ctx context.Context, user models.User) (int, error)
}

type userUsecase struct {
	tgBot  *tgbotapi.BotAPI
	psql   repository.IPostgresqlRepository
	rds    repository.IRedisRepository
	logger logrus.Logger
}

func NewUserUsecase(
	tgBot *tgbotapi.BotAPI,
	pr repository.IPostgresqlRepository,
	rr repository.IRedisRepository,
	logger logrus.Logger,
) IUserUsecase {
	return &userUsecase{
		tgBot:  tgBot,
		psql:   pr,
		rds:    rr,
		logger: logger,
	}
}

func (uu *userUsecase) CheckStatus(ctx context.Context) (int, error) {
	return uu.psql.CheckStatus(ctx)
}

func (uu *userUsecase) GetUserByCookie(ctx context.Context, cookie string) (models.User, int, error) {
	userLogin, err := uu.rds.CheckSession(ctx, cookie)
	if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.GetUserByCookie: failed to check session in redis: %w", err)
	}

	if userLogin == "" {
		return models.User{},
			http.StatusForbidden,
			fmt.Errorf("UserUsercase.GetUserByCookie: user not authorized")
	}

	user, err := uu.psql.GetUserByLogin(ctx, userLogin)
	if err == pgx.ErrNoRows {
		return models.User{},
			http.StatusNotFound,
			fmt.Errorf("UserUsecase.GetUserByCookie: user with same email not found")
	} else if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.GetUserByCookie: failed to check email in db with err: %s", err)
	}

	return user, http.StatusOK, nil
}

func (uu *userUsecase) GetUserByNickname(ctx context.Context, nickname string) (models.User, int, error) {
	user, err := uu.psql.GetUserByNickname(ctx, nickname)
	if err == pgx.ErrNoRows {
		return models.User{},
			http.StatusNotFound,
			fmt.Errorf("UserUsecase.GetUserByNickname: user with same nickname not found")
	} else if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.GetUserByNickname: failed to check nickname in db with err: %s", err)
	}

	return user, http.StatusOK, nil
}

func (uu *userUsecase) SignupUser(
	ctx context.Context,
	credentials models.SignupCredentials,
) (models.User, string, int, error) {
	// checking that user is new
	_, err := uu.psql.GetUserByEmailOrNickname(ctx, credentials.Email, credentials.Nickname)
	if err == nil {
		return models.User{},
			"",
			http.StatusConflict,
			fmt.Errorf("UserUsecase.SignupUser: user with same email or nickname already exists")
	} else if err != nil && err != pgx.ErrNoRows {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to check user email and nickname in db with err: %s", err)
	}

	// creating new cookie value
	sessionID, err := uuid.NewRandom()
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to create sessionID: %w", err)
	}

	// setting cookie-email in redis
	err = uu.rds.CreateSession(ctx, sessionID.String(), credentials.Email, config.ExpirationCookieTime)
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to create sessionID: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() { uu.rds.DeleteSession(ctx, sessionID.String()) })

	// hashing password
	hashedPswd, err := hasher.HashAndSalt(credentials.Password)
	if err != nil {
		// rollback
		rb.Run()

		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to hash password: %w", err)
	}
	credentials.Password = hashedPswd
	credentials.Avatar = fmt.Sprintf("%s/%s", consts.AvatarsBaseFolderPath, consts.DefaultAvatarFileName)

	// TODO: чтото с генерацией никнейма
	// generate nickname if it's empty
	// if credentials.Nickname == "" {
	// 	credentials.Nickname = uuid.NewString()
	// }

	// inserting user in psql
	createdUser, err := uu.psql.AddUser(ctx, credentials)
	if err != nil {
		// rollback
		rb.Run()

		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to save user in db: %w", err)
	}

	return createdUser, sessionID.String(), http.StatusOK, nil
}

func (uu *userUsecase) LoginUser(
	ctx context.Context,
	credentials models.LoginCredentials,
) (models.User, string, int, error) {
	// creating new cookie value
	sessionID, err := uuid.NewRandom()
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to create sessionID: %w", err)
	}

	// setting cookie-email in redis
	err = uu.rds.CreateSession(ctx, sessionID.String(), credentials.Login, config.ExpirationCookieTime)
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to create sessionID: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() { uu.rds.DeleteSession(ctx, sessionID.String()) })

	// finding user in psql
	user, err := uu.psql.GetUserByLogin(ctx, credentials.Login)
	if err == pgx.ErrNoRows {
		return models.User{},
			"",
			http.StatusNotFound,
			fmt.Errorf("UserUsecase.LoginUser: user with same email or nickname not found")
	} else if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to check email or nickname in db with err: %w", err)
	}

	// checking password
	isEqual, err := hasher.ComparePasswords(user.Password, credentials.Password)
	if err != nil {
		// rollback
		rb.Run()

		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to compare passwords: %w", err)
	}
	if !isEqual {
		// rollback
		rb.Run()

		return models.User{},
			"",
			http.StatusForbidden,
			fmt.Errorf("UserUsecase.LoginUser: user not authorized")
	}

	return user, sessionID.String(), http.StatusOK, nil
}

func (uu *userUsecase) LogoutUser(ctx context.Context, cookieValue string) (int, error) {
	// deleting session from redis
	err := uu.rds.DeleteSession(ctx, cookieValue)
	if err != nil {
		return http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LogoutUser: failed to delete session from redis: %w", err)
	}

	return http.StatusOK, nil
}

func (uu *userUsecase) UpdateUser(ctx context.Context, newUserData models.UpdateUserReq) (models.User, int, error) {
	updatedUser, err := uu.psql.UpdateUser(ctx, newUserData)
	if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.UpdateUser: failed to update user in db: %w", err)
	}

	return updatedUser, http.StatusOK, nil
}

func (uu *userUsecase) DeleteUser(ctx context.Context, user models.User, cookieValue string) (int, error) {
	// deleting session from redis
	err := uu.rds.DeleteSession(ctx, cookieValue)
	if err != nil {
		return http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.DeleteUser: failed to delete session from redis: %w", err)
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() { uu.rds.CreateSession(ctx, cookieValue, user.Email, config.ExpirationCookieTime) })

	// deleteing user in psql
	err = uu.psql.DeleteUser(ctx, user.ID)
	if err != nil {
		// rollback
		rb.Run()

		return http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.DeleteUser: failed to delete user in db: %w", err)
	}

	return http.StatusOK, nil
}

type UpdateAvatarArgs struct {
	User       models.User
	FileHeader *multipart.FileHeader
	File       multipart.File
}

func (uu *userUsecase) UpdateAvatar(
	ctx context.Context,
	args UpdateAvatarArgs,
) (models.User, int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	oldAvatar := args.User.Avatar

	buf := make([]byte, args.FileHeader.Size)
	_, err := args.File.Read(buf)
	if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.UpdateAvatar: failed to read file: %w", err)
	}

	fileType := http.DetectContentType(buf)
	if !file_utils.IsEnabledFileType(fileType) {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.UpdateAvatar: not enabled file type")
	}

	folderNameByte := sha1.New().Sum([]byte(args.User.Email))
	folderName := fmt.Sprintf(hex.EncodeToString(folderNameByte))

	// saving avatar file
	avatarFileName := file_utils.GenerateFileName("avatar", consts.FileExt)
	avatarFolderPath := fmt.Sprintf("%s/%s", consts.AvatarsBaseFolderPath, folderName)
	avatarFilePath := fmt.Sprintf("%s/%s/%s", consts.AvatarsBaseFolderPath, folderName, avatarFileName)

	err = file_utils.SaveFile(avatarFolderPath, avatarFilePath, buf)
	if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.UpdateAvatar: failed to save avatar's file: %w", err)
	}

	rb.Add(func() {
		err := os.Remove(avatarFilePath)
		if err != nil {
			uu.logger.Errorf("UserUsecase.UpdateAvatar: failed to rollback creating of avatar's file: %w", err)
		}
	})

	updatedUser, err := uu.psql.UpdateAvatar(ctx, args.User.ID, avatarFilePath)
	if err != nil {
		rb.Run()

		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.UpdateAvatar: failed to update user's avatar in db: %w", err)
	}

	rb.Add(func() {
		_, err := uu.psql.UpdateAvatar(ctx, args.User.ID, oldAvatar)
		if err != nil {
			uu.logger.Errorf("UserUsecase.UpdateAvatar: failed to rollback updating of user's avatar in db: %w", err)
		}
	})

	// if user already have avatar deleting
	if oldAvatar != "" {
		err := os.Remove(oldAvatar)
		if err != nil {
			rb.Run()

			return models.User{},
				http.StatusInternalServerError,
				fmt.Errorf("UserUsecase.UpdateAvatar: failed to delete old avatar's file: %w", err)
		}
	}

	return updatedUser, http.StatusOK, nil
}

func (uu *userUsecase) DeleteAvatar(ctx context.Context, user models.User) (models.User, int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	oldAvatar := user.Avatar

	defaultAvatarFilePath := fmt.Sprintf("%s/%s", consts.DefaultsBaseFolderPath, consts.DefaultAvatarFileName)

	updatedUser, err := uu.psql.UpdateAvatar(ctx, user.ID, defaultAvatarFilePath)
	if err != nil {
		rb.Run()

		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.DeleteAvatar: failed to update user's avatar in db: %w", err)
	}

	rb.Add(func() {
		_, err := uu.psql.UpdateAvatar(ctx, user.ID, oldAvatar)
		if err != nil {
			uu.logger.Errorf("UserUsecase.DeleteAvatar: failed to rollback updating of user's avatar in db: %w", err)
		}
	})

	// deleting previous avatar
	if oldAvatar != "" {
		err := os.Remove(oldAvatar)
		if err != nil {
			rb.Run()

			return models.User{},
				http.StatusInternalServerError,
				fmt.Errorf("UserUsecase.DeleteAvatar: failed to delete old avatar's file: %w", err)
		}
	}

	return updatedUser, http.StatusOK, nil
}

// TODO: move into separate usecase
func (uu *userUsecase) BecomeStylist(ctx context.Context, user models.User) (int, error) {
	ctx, rb := rollback.NewCtxRollback(ctx)

	_, err := uu.psql.GetUserStylistRequestByUID(ctx, user.ID)
	if err != nil {
		if err != pgx.ErrNoRows {
			return http.StatusInternalServerError,
				fmt.Errorf("UserUsecase.BecomeStylist: failed to get stylist request from db: %w", err)
		}
	}

	// user already have stylist request
	if err == nil {
		return http.StatusConflict,
			fmt.Errorf("UserUsecase.BecomeStylist: user already have stylist request")
	}

	createdReq, err := uu.psql.AddStylistRequest(ctx, user.ID)
	if err != nil {
		return http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.BecomeStylist: failed to save stylist request in db: %w", err)
	}

	rb.Add(func() {
		_, err := uu.psql.DeleteStylistRequestByID(ctx, createdReq.ID)
		if err != nil {
			uu.logger.Errorf("UserUsecase.DeleteAvatar: failed to rollback saving of stylist request in db: %w", err)
		}
	})

	msgText := fmt.Sprintf(consts.StylistRequestMsg, createdReq.ID, user.Email, user.Nickname, user.Name, user.Age, user.Desc)
	msg := tgbotapi.NewMessage(config.TgBotAPIToken.AdminChatID, msgText)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Принять", fmt.Sprintf("%d,%d", consts.TGBotResponseAccept, createdReq.ID)),
			tgbotapi.NewInlineKeyboardButtonData("Отклонить", fmt.Sprintf("%d,%d", consts.TGBotResponseReject, createdReq.ID)),
		),
	)

	if _, err := uu.tgBot.Send(msg); err != nil {
		rb.Run()

		return http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.BecomeStylist: failed to send msg in tg chat: %w", err)
	}

	return http.StatusOK, nil
}

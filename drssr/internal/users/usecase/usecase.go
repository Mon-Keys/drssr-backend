package usecase

import (
	"context"
	"drssr/config"
	"drssr/internal/models"
	"drssr/internal/pkg/hasher"
	"drssr/internal/pkg/rollback"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"

	"drssr/internal/users/repository"
)

type IUserUsecase interface {
	GetUserByCookie(ctx context.Context, cookie string) (models.User, int, error)
	GetUserByNickname(ctx context.Context, nickname string) (models.User, int, error)
	SignupUser(ctx context.Context, credentials models.SignupCredentials) (models.User, string, int, error)
	LoginUser(ctx context.Context, credentials models.LoginCredentials) (models.User, string, int, error)
}

type userUsecase struct {
	psql   repository.IPostgresqlRepository
	rds    repository.IRedisRepository
	logger logrus.Logger
}

func NewUserUsecase(
	pr repository.IPostgresqlRepository,
	rr repository.IRedisRepository,
	logger logrus.Logger,
) IUserUsecase {
	return &userUsecase{
		psql:   pr,
		rds:    rr,
		logger: logger,
	}
}

func (uu *userUsecase) GetUserByCookie(ctx context.Context, cookie string) (models.User, int, error) {
	userLogin, err := uu.rds.CheckSession(ctx, cookie)
	if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.CheckSession: failed to check session in redis")
	}

	if userLogin == "" {
		return models.User{},
			http.StatusForbidden,
			fmt.Errorf("UserUsercase.CheckSession: user not authorized")
	}

	var user models.User
	if strings.Contains(userLogin, "@") {
		user, err = uu.psql.GetUserByEmail(ctx, userLogin)
	} else {
		user, err = uu.psql.GetUserByNickname(ctx, userLogin)
	}
	if err == pgx.ErrNoRows {
		return models.User{},
			http.StatusNotFound,
			fmt.Errorf("UserUsecase.CheckSession: user with same email or nickname not found")
	} else if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.CheckSession: failed to check email or nickname in db with err: %s", err)
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
	// creating new cookie value
	sessionID, err := uuid.NewRandom()
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to create sessionID")
	}

	// setting cookie-email in redis
	err = uu.rds.CreateSession(ctx, sessionID.String(), credentials.Email, config.ExpirationCookieTime)
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.SignupUser: failed to create sessionID")
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
			fmt.Errorf("UserUsecase.SignupUser: failed to hash password")
	}
	credentials.Password = hashedPswd

	// TODO: сделать генерацию никнейма
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
			fmt.Errorf("UserUsecase.SignupUser: failed to save user in db")
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
			fmt.Errorf("UserUsecase.LoginUser: failed to create sessionID")
	}

	// setting cookie-email in redis
	err = uu.rds.CreateSession(ctx, sessionID.String(), credentials.Login, config.ExpirationCookieTime)
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to create sessionID")
	}

	ctx, rb := rollback.NewCtxRollback(ctx)
	rb.Add(func() { uu.rds.DeleteSession(ctx, sessionID.String()) })

	// finding user in psql
	var user models.User
	if strings.Contains(credentials.Login, "@") {
		user, err = uu.psql.GetUserByEmail(ctx, credentials.Login)
	} else {
		user, err = uu.psql.GetUserByNickname(ctx, credentials.Login)
	}
	if err == pgx.ErrNoRows {
		return models.User{},
			"",
			http.StatusNotFound,
			fmt.Errorf("UserUsecase.LoginUser: user with same email or nickname not found")
	} else if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to check email or nickname in db with err: %s", err)
	}

	// checking password
	isEqual, err := hasher.ComparePasswords(user.Password, credentials.Password)
	if err != nil {
		// rollback
		rb.Run()

		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.LoginUser: failed to compare passwords")
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

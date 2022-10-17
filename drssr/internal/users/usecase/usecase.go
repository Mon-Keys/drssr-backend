package usecase

import (
	"context"
	"drssr/config"
	"drssr/internal/models"
	"drssr/internal/pkg/hasher"
	"drssr/internal/pkg/rollback"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"

	"drssr/internal/users/repository"
)

type IUserUsecase interface {
	GetUserByCookie(ctx context.Context, cookie string) (models.User, int, error)
	AddUser(ctx context.Context, credentials models.SignupCredentials) (models.User, string, int, error)
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
	userEmail, err := uu.rds.CheckSession(ctx, cookie)
	if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.CheckSession: failed to check session in redis")
	}

	if userEmail == "" {
		return models.User{},
			http.StatusForbidden,
			fmt.Errorf("UserUsercase.CheckSession: user not authorized")
	}

	user, err := uu.psql.GetUserByEmail(ctx, userEmail)
	if err == pgx.ErrNoRows {
		return models.User{},
			http.StatusNotFound,
			fmt.Errorf("UserUsecase.CheckSession: user with same email not found")
	} else if err != nil {
		return models.User{},
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.CheckSession: failed to check email in db with err: %s", err)
	}

	return user, http.StatusOK, nil
}

func (uu *userUsecase) AddUser(
	ctx context.Context,
	credentials models.SignupCredentials,
) (models.User, string, int, error) {
	// creating new cookie value
	sessionID, err := uuid.NewRandom()
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.AddUser: failed to create sessionID")
	}

	// setting cookie-email in redis
	err = uu.rds.CreateSession(ctx, sessionID.String(), credentials.Email, config.ExpirationCookieTime)
	if err != nil {
		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.AddUser: failed to create sessionID")
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
			fmt.Errorf("UserUsecase.AddUser: failed to hash password")
	}
	credentials.Password = hashedPswd

	// inserting user in psql
	createdUser, err := uu.psql.AddUser(ctx, credentials)
	if err != nil {
		// rollback
		rb.Run()

		return models.User{},
			"",
			http.StatusInternalServerError,
			fmt.Errorf("UserUsecase.AddUser: failed to save user in db")
	}

	return createdUser, sessionID.String(), http.StatusOK, nil
}

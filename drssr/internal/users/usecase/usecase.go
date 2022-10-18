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
	GetUserByNickname(ctx context.Context, nickname string) (models.User, int, error)
	SignupUser(ctx context.Context, credentials models.SignupCredentials) (models.User, string, int, error)
	LoginUser(ctx context.Context, credentials models.LoginCredentials) (models.User, string, int, error)
	LogoutUser(ctx context.Context, email string) (int, error)
	UpdateUser(ctx context.Context, newUserData models.UpdateUserReq) (models.User, int, error)
	DeleteUser(ctx context.Context, user models.User, cookieValue string) (int, error)
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

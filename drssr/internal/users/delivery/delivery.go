package delivery

import (
	"drssr/config"
	"drssr/internal/models"
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/ioutils"
	middleware "drssr/internal/pkg/middlewares"
	"drssr/internal/users/usecase"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type UserDelivery struct {
	userUseCase usecase.IUserUsecase
	logger      logrus.Logger
}

func SetUserRouting(
	router *mux.Router,
	uu usecase.IUserUsecase,
	authMw middleware.AuthMiddleware,
	logger logrus.Logger,
) {
	userDelivery := &UserDelivery{
		userUseCase: uu,
		logger:      logger,
	}

	// public API
	userPublicAPI := router.PathPrefix("/api/v1/public/users").Subrouter()

	userPublicAPI.Use(middleware.WithRequestID, middleware.WithJSON)

	userPublicAPI.HandleFunc("/signup", userDelivery.signup).Methods(http.MethodPost)
	userPublicAPI.HandleFunc("/login", userDelivery.login).Methods(http.MethodPost)
	userPublicAPI.HandleFunc("", userDelivery.getSomeUser).Methods(http.MethodGet)

	// private API
	userPrivateAPI := router.PathPrefix("/api/v1/private/users").Subrouter()

	userPrivateAPI.Use(middleware.WithRequestID, middleware.WithJSON, authMw.WithAuth)

	userPrivateAPI.HandleFunc("", userDelivery.getUser).Methods(http.MethodGet)
	userPrivateAPI.HandleFunc("", userDelivery.updateUser).Methods(http.MethodPut)
	userPrivateAPI.HandleFunc("", userDelivery.deleteUser).Methods(http.MethodDelete)
	userPrivateAPI.HandleFunc("/logout", userDelivery.logout).Methods(http.MethodDelete)
}

// public

func (ud *UserDelivery) signup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})

	var credentials models.SignupCredentials
	err := ioutils.ReadJSON(r, &credentials)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if credentials.Email == "" || credentials.Password == "" || credentials.Nickname == "" || credentials.BirthDate == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	createdUser, sessionID, status, err := ud.userUseCase.SignupUser(ctx, credentials)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to signup user: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  sessionID,
		MaxAge: int(config.ExpirationCookieTime),
		Path:   "/api/v1",
	}

	http.SetCookie(w, cookie)
	ioutils.Send(w, status, createdUser)
}

func (ud *UserDelivery) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})

	var credentials models.LoginCredentials
	err := ioutils.ReadJSON(r, &credentials)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if credentials.Login == "" || credentials.Password == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	user, sessionID, status, err := ud.userUseCase.LoginUser(ctx, credentials)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to login user: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  sessionID,
		MaxAge: int(config.ExpirationCookieTime),
		Path:   "/api/v1",
	}

	http.SetCookie(w, cookie)
	ioutils.Send(w, status, user)
}

func (ud *UserDelivery) getSomeUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})

	nickname := r.URL.Query().Get("nickname")
	if nickname == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	user, status, err := ud.userUseCase.GetUserByNickname(ctx, nickname)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to get user by nickname: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, user)
}

// private

func (ud *UserDelivery) getUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ud.logger = *ud.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	ioutils.Send(w, http.StatusOK, user)
}

func (ud *UserDelivery) updateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ud.logger = *ud.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	var newUserData models.UpdateUserReq
	err := ioutils.ReadJSON(r, &newUserData)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	updateUser, status, err := ud.userUseCase.UpdateUser(ctx, newUserData)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", http.StatusInternalServerError).Errorf("Failed to update user: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, updateUser)
}

func (ud *UserDelivery) deleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ud.logger = *ud.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	cookieToken, err := r.Cookie("session-id")
	if err != nil {
		logger.WithField("status", http.StatusInternalServerError).Errorf("Failed to get cookie from request")
		ioutils.SendDefaultError(w, http.StatusInternalServerError)
		return
	}

	status, err := ud.userUseCase.DeleteUser(ctx, *user, cookieToken.Value)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", http.StatusInternalServerError).Errorf("Failed to delete user: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  "",
		MaxAge: -1,
		Path:   "/api/v1",
	}

	http.SetCookie(w, cookie)
	ioutils.SendWithoutBody(w, status)
}

func (ud *UserDelivery) logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ud.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ud.logger = *ud.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	cookieToken, err := r.Cookie("session-id")
	if err != nil {
		logger.WithField("status", http.StatusInternalServerError).Errorf("Failed to get cookie from request")
		ioutils.SendDefaultError(w, http.StatusInternalServerError)
		return
	}

	status, err := ud.userUseCase.LogoutUser(ctx, cookieToken.Value)
	if err != nil || status != http.StatusOK {
		logger.WithField(
			"status",
			http.StatusInternalServerError,
		).Errorf("Failed to delete cookie value from redis: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	cookie := &http.Cookie{
		Name:   "session-id",
		Value:  "",
		MaxAge: -1,
		Path:   "/api/v1",
	}

	http.SetCookie(w, cookie)
	ioutils.SendWithoutBody(w, status)
}

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
	logger logrus.Logger,
) {
	userDelivery := &UserDelivery{
		userUseCase: uu,
		logger:      logger,
	}

	// public API
	userPublicAPI := router.PathPrefix("/api/v1/users/public").Subrouter()

	userPublicAPI.Use(middleware.WithRequestID, middleware.WithJSON)

	userPublicAPI.HandleFunc("/signup", userDelivery.signup).Methods(http.MethodPost)
	userPublicAPI.HandleFunc("/login", userDelivery.login).Methods(http.MethodPost)
	userPublicAPI.HandleFunc("/", userDelivery.getSomeUser).Methods(http.MethodGet)

	authMw := middleware.NewAuthMiddleware(uu, logger)

	// private API
	userPrivateAPI := router.PathPrefix("/api/v1/users/private").Subrouter()

	userPrivateAPI.Use(middleware.WithRequestID, middleware.WithJSON, authMw.WithAuth)

	userPrivateAPI.HandleFunc("/", userDelivery.getUser).Methods(http.MethodGet)
	userPrivateAPI.HandleFunc("/", userDelivery.updateUser).Methods(http.MethodPut)
	userPrivateAPI.HandleFunc("/", userDelivery.deleteUser).Methods(http.MethodDelete)
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
		logger.WithField("status", http.StatusBadRequest).Errorf("failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if credentials.Email == "" || credentials.Password == "" || credentials.Nickname == "" || credentials.BirthDate == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	createdUser, sessionID, status, err := ud.userUseCase.AddUser(ctx, credentials)

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
}

func (ud *UserDelivery) getSomeUser(w http.ResponseWriter, r *http.Request) {
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
		logger.WithField("status", http.StatusForbidden).Errorf("failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

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
		logger.WithField("status", http.StatusForbidden).Errorf("failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}
	// TODO registration
}

func (ud *UserDelivery) deleteUser(w http.ResponseWriter, r *http.Request) {
}

func (ud *UserDelivery) logout(w http.ResponseWriter, r *http.Request) {
}

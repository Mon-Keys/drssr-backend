package delivery

import (
	"drssr/internal/looks/usecase"
	"drssr/internal/models"
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/ioutils"
	middleware "drssr/internal/pkg/middlewares"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type LooksDelivery struct {
	looksUseCase usecase.ILooksUsecase
	logger       logrus.Logger
}

func SetLooksRouting(
	router *mux.Router,
	cu usecase.ILooksUsecase,
	authMw middleware.AuthMiddleware,
	logger logrus.Logger,
) {
	clothesDelivery := &LooksDelivery{
		looksUseCase: cu,
		logger:       logger,
	}

	// clothes API
	looksPrivateAPI := router.PathPrefix("/api/v1/private/looks").Subrouter()
	looksPrivateAPI.Use(middleware.WithRequestID, middleware.WithJSON, authMw.WithAuth)

	looksPrivateAPI.HandleFunc("", clothesDelivery.addLook).Methods(http.MethodPost)
	looksPrivateAPI.HandleFunc("", clothesDelivery.getLook).Methods(http.MethodGet)
	looksPrivateAPI.HandleFunc("", clothesDelivery.updateLook).Methods(http.MethodPut)
	// looksPrivateAPI.HandleFunc("", clothesDelivery).Methods(http.MethodDelete)

	looksPublicAPI := router.PathPrefix("/api/v1/public/looks").Subrouter()
	looksPublicAPI.Use(middleware.WithRequestID, middleware.WithJSON)

	// looksPublicAPI.HandleFunc("/all", clothesDelivery).Methods(http.MethodGet)
}

func (ld *LooksDelivery) addLook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ld.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ld.logger = *ld.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	var look models.Look
	err := ioutils.ReadJSON(r, &look)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if look.Img == "" || look.CreatorID == 0 || len(look.Clothes) == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	createdLook, status, err := ld.looksUseCase.AddLook(ctx, look)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to save look: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, createdLook)
}

func (ld *LooksDelivery) getLook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ld.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ld.logger = *ld.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	lidParam := r.URL.Query().Get("id")
	if lidParam == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}
	lid, err := strconv.ParseUint(lidParam, 10, 64)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if lid == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	look, status, err := ld.looksUseCase.GetLookByID(ctx, lid, user.ID)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to save look: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, look)
}

func (ld *LooksDelivery) updateLook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := ld.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	ld.logger = *ld.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	var look models.Look
	err := ioutils.ReadJSON(r, &look)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	lidParam := r.URL.Query().Get("id")
	if lidParam == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}
	lid, err := strconv.ParseUint(lidParam, 10, 64)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if look.Img == "" || look.CreatorID == 0 || len(look.Clothes) == 0 || lid == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	updatedLook, status, err := ld.looksUseCase.UpdateLook(ctx, look, lid, user.ID)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to save look: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, updatedLook)
}

// func (cd *ClothesDelivery) getAllClothes(w http.ResponseWriter, r *http.Request) {
// 	var err error

// 	ctx := r.Context()
// 	reqID := ctx_utils.GetReqID(ctx)
// 	logger := cd.logger.WithFields(logrus.Fields{
// 		"url":    r.URL,
// 		"req_id": reqID,
// 	})

// 	queryParams := r.URL.Query()

// 	limitStr := queryParams.Get("limit")
// 	limitInt := 0
// 	if limitStr != "" {
// 		limitInt, err = strconv.Atoi(limitStr)
// 		if err != nil || limitInt < 0 {
// 			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse limit: %w", err)
// 			ioutils.SendDefaultError(w, http.StatusBadRequest)
// 			return
// 		}
// 	}
// 	offsetStr := queryParams.Get("offset")
// 	offsetInt := 0
// 	if offsetStr != "" {
// 		offsetInt, err = strconv.Atoi(offsetStr)
// 		if err != nil || offsetInt < 0 {
// 			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse offset: %w", err)
// 			ioutils.SendDefaultError(w, http.StatusBadRequest)
// 			return
// 		}
// 	}

// 	allClothes, status, err := cd.clothesUseCase.GetAllClothes(ctx, limitInt, offsetInt)
// 	if err != nil || status != http.StatusOK {
// 		logger.WithField("status", status).Errorf("Failed to get clothes: %w", err)
// 		ioutils.SendDefaultError(w, status)
// 		return
// 	}

// 	ioutils.Send(w, status, allClothes)
// }

// func (cd *ClothesDelivery) getUsersClothes(w http.ResponseWriter, r *http.Request) {
// 	var err error

// 	ctx := r.Context()
// 	reqID := ctx_utils.GetReqID(ctx)
// 	logger := cd.logger.WithFields(logrus.Fields{
// 		"url":    r.URL,
// 		"req_id": reqID,
// 	})
// 	user := ctx_utils.GetUser(ctx)
// 	if user == nil {
// 		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
// 		ioutils.SendDefaultError(w, http.StatusForbidden)
// 		return
// 	}

// 	cd.logger = *cd.logger.WithFields(logrus.Fields{
// 		"user": user.Email,
// 	}).Logger

// 	queryParams := r.URL.Query()

// 	limitStr := queryParams.Get("limit")
// 	limitInt := 0
// 	if limitStr != "" {
// 		limitInt, err = strconv.Atoi(limitStr)
// 		if err != nil || limitInt < 0 {
// 			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse limit: %w", err)
// 			ioutils.SendDefaultError(w, http.StatusBadRequest)
// 			return
// 		}
// 	}
// 	offsetStr := queryParams.Get("offset")
// 	offsetInt := 0
// 	if offsetStr != "" {
// 		offsetInt, err = strconv.Atoi(offsetStr)
// 		if err != nil || offsetInt < 0 {
// 			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse offset: %w", err)
// 			ioutils.SendDefaultError(w, http.StatusBadRequest)
// 			return
// 		}
// 	}

// 	allClothes, status, err := cd.clothesUseCase.GetUsersClothes(ctx, limitInt, offsetInt, user.ID)
// 	if err != nil || status != http.StatusOK {
// 		logger.WithField("status", status).Errorf("Failed to get user's clothes: %w", err)
// 		ioutils.SendDefaultError(w, status)
// 		return
// 	}

// 	ioutils.Send(w, status, allClothes)
// }

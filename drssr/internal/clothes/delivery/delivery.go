package delivery

import (
	"drssr/internal/clothes/usecase"
	"drssr/internal/models"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/file_utils"
	"drssr/internal/pkg/ioutils"
	middleware "drssr/internal/pkg/middlewares"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type ClothesDelivery struct {
	clothesUseCase usecase.IClothesUsecase
	logger         logrus.Logger
}

func SetClothesRouting(
	router *mux.Router,
	cu usecase.IClothesUsecase,
	authMw middleware.AuthMiddleware,
	logger logrus.Logger,
) {
	clothesDelivery := &ClothesDelivery{
		clothesUseCase: cu,
		logger:         logger,
	}

	// private API
	clothesPrivateAPI := router.PathPrefix("/api/v1/private/clothes").Subrouter()
	clothesPrivateAPI.Use(middleware.WithRequestID, middleware.WithJSON, authMw.WithAuth)

	clothesPrivateAPI.HandleFunc("", clothesDelivery.addClothes).Methods(http.MethodPost)
	clothesPrivateAPI.HandleFunc("", clothesDelivery.updateClothes).Methods(http.MethodPut)
	clothesPrivateAPI.HandleFunc("", clothesDelivery.getUsersClothes).Methods(http.MethodGet)
	clothesPrivateAPI.HandleFunc("", clothesDelivery.deleteClothes).Methods(http.MethodDelete)

	// public API
	clothesPublicAPI := router.PathPrefix("/api/v1/public/clothes").Subrouter()
	clothesPublicAPI.Use(middleware.WithRequestID, middleware.WithJSON)

	clothesPublicAPI.HandleFunc("/all", clothesDelivery.getAllClothes).Methods(http.MethodGet)
}

func (cd *ClothesDelivery) addClothes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := cd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	cd.logger = *cd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	r.Body = http.MaxBytesReader(w, r.Body, consts.MaxUploadFileSize)
	if err := r.ParseMultipartForm(consts.MaxUploadFileSize); err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse multipart/form-data request: %w", err)
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	file, fileHeader, status, err := file_utils.OpenFileFromReq(r, "file")
	if err != nil {
		logger.WithField("status", status).Errorf("Opening file error: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	createdClothes, status, err := cd.clothesUseCase.AddFile(ctx, usecase.AddFileArgs{
		UID:        user.ID,
		UserEmail:  user.Email,
		FileHeader: fileHeader,
		File:       *file,
	})
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to add file: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, createdClothes)
}

func (cd *ClothesDelivery) updateClothes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := cd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	cd.logger = *cd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	clothesIDStr := r.URL.Query().Get("id")
	if clothesIDStr == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid clothes ID query param")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	clothesID, err := strconv.ParseUint(clothesIDStr, 10, 64)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse url query")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if clothesID == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid clothes ID")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	var newClothesData models.Clothes
	err = ioutils.ReadJSON(r, &newClothesData)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	newClothesData.ID = clothesID

	updatedClothes, status, err := cd.clothesUseCase.UpdateClothes(ctx, user.ID, newClothesData)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to update clothes: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, updatedClothes)
}

func (cd *ClothesDelivery) deleteClothes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := cd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	cd.logger = *cd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	clothesIDStr := r.URL.Query().Get("id")
	if clothesIDStr == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid clothes ID query param")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	clothesID, err := strconv.ParseUint(clothesIDStr, 10, 64)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse url query")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if clothesID == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid clothes ID")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	status, err := cd.clothesUseCase.DeleteClothes(ctx, user.ID, clothesID)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to delete clothes: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.SendWithoutBody(w, status)
}

func (cd *ClothesDelivery) getAllClothes(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := cd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})

	queryParams := r.URL.Query()

	limitStr := queryParams.Get("limit")
	limitInt := 0
	if limitStr != "" {
		limitInt, err = strconv.Atoi(limitStr)
		if err != nil || limitInt < 0 || limitInt > consts.GetClothesLimit {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse limit: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}
	offsetStr := queryParams.Get("offset")
	offsetInt := 0
	if offsetStr != "" {
		offsetInt, err = strconv.Atoi(offsetStr)
		if err != nil || offsetInt < 0 {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse offset: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}

	allClothes, status, err := cd.clothesUseCase.GetAllClothes(ctx, limitInt, offsetInt)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to get clothes: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, allClothes)
}

func (cd *ClothesDelivery) getUsersClothes(w http.ResponseWriter, r *http.Request) {
	var err error

	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := cd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	cd.logger = *cd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	queryParams := r.URL.Query()

	limitStr := queryParams.Get("limit")
	limitInt := 0
	if limitStr != "" {
		limitInt, err = strconv.Atoi(limitStr)
		if err != nil || limitInt < 0 || limitInt > consts.GetClothesLimit {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse limit: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}
	offsetStr := queryParams.Get("offset")
	offsetInt := 0
	if offsetStr != "" {
		offsetInt, err = strconv.Atoi(offsetStr)
		if err != nil || offsetInt < 0 {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse offset: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}

	allClothes, status, err := cd.clothesUseCase.GetUsersClothes(ctx, limitInt, offsetInt, user.ID)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to get user's clothes: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, allClothes)
}

package delivery

import (
	"drssr/internal/clothes/usecase"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/ctx_utils"
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

	// clothes API
	clothesPrivateAPI := router.PathPrefix("/api/v1/clothes").Subrouter()
	clothesPrivateAPI.Use(middleware.WithRequestID, middleware.WithJSON, authMw.WithAuth)

	clothesPublicAPI := router.PathPrefix("/api/v1/clothes").Subrouter()
	clothesPublicAPI.Use(middleware.WithRequestID, middleware.WithJSON)

	// clothesPrivateAPI.HandleFunc("", ClothesDelivery.getUser).Methods(http.MethodGet)
	clothesPrivateAPI.HandleFunc("", clothesDelivery.addClothes).Methods(http.MethodPost)
	clothesPrivateAPI.HandleFunc("", clothesDelivery.getUsersClothes).Methods(http.MethodGet)
	clothesPublicAPI.HandleFunc("/all", clothesDelivery.getAllClothes).Methods(http.MethodGet)
	// clothesPrivateAPI.HandleFunc("", ClothesDelivery.deleteUser).Methods(http.MethodDelete)

	// clothesPublicAPI := router.PathPrefix("/api/v1/clothes").Subrouter()

	// clothesPublicAPI.Use(middleware.WithRequestID, middleware.Wclothes
	// clothesPublicAPI.HandleFunc("/all", ClothesDelivery.logout).Methods(http.MethodGet)
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
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse multipart/form-data request")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	clothesSex := r.FormValue("sex")
	clothesBrand := r.FormValue("brand")

	// we upload only 1 file
	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("No file in request")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}
	fileHeader := files[0]
	if fileHeader.Size > consts.MaxUploadFileSize {
		logger.WithField("status", http.StatusBadRequest).Errorf("File is too big")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		logger.WithField("status", http.StatusInternalServerError).Errorf("Failed to open file")
		ioutils.SendDefaultError(w, http.StatusInternalServerError)
		return
	}

	createdClothes, status, err := cd.clothesUseCase.AddFile(ctx, usecase.AddFileArgs{
		UID:          user.ID,
		FileHeader:   fileHeader,
		File:         file,
		ClothesBrand: clothesBrand,
		ClothesSex:   clothesSex,
	})
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to add file: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, createdClothes)
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
		if err != nil || limitInt < 0 {
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
		if err != nil || limitInt < 0 {
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

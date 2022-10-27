package delivery

import (
	"drssr/internal/looks/usecase"
	"drssr/internal/pkg/common"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/ioutils"
	middleware "drssr/internal/pkg/middlewares"
	"net/http"
	"strconv"
	"strings"

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
	// looksPrivateAPI.HandleFunc("", clothesDelivery).Methods(http.MethodGet)
	// looksPrivateAPI.HandleFunc("", clothesDelivery).Methods(http.MethodPut)
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

	r.Body = http.MaxBytesReader(w, r.Body, consts.MaxUploadFileSize)
	if err := r.ParseMultipartForm(consts.MaxUploadFileSize); err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse multipart/form-data request: %w", err)
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	clothesStr := r.FormValue("clothes")
	clothesStr = strings.Trim(clothesStr, "[]")
	clothesSplitedStr := strings.Split(clothesStr, ",")
	if len(clothesSplitedStr) == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Empty closes array")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	clothes := make([]uint64, len(clothesSplitedStr), len(clothesSplitedStr))
	for i, clothesIDStr := range clothesSplitedStr {
		clothesID, err := strconv.ParseUint(clothesIDStr, 10, 64)
		if err != nil {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse clothes array: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
		clothes[i] = clothesID
	}

	description := r.FormValue("description")

	file, fileHeader, status, err := common.OpenFileFromReq(r, "file")
	if err != nil {
		logger.WithField("status", status).Errorf("Opening file error: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	previewFile, previewFileHeader, status, err := common.OpenFileFromReq(r, "preview_file")
	if err != nil {
		logger.WithField("status", status).Errorf("Opening preview file error: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	createdClothes, status, err := ld.looksUseCase.AddLook(ctx, usecase.AddLookArgs{
		UID:               user.ID,
		FileHeader:        fileHeader,
		File:              *file,
		PreviewFileHeader: previewFileHeader,
		PreviewFile:       *previewFile,
		Clothes:           clothes,
		Description:       description,
	})
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to add file: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, createdClothes)
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

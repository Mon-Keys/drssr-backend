package delivery

import (
	"drssr/internal/clothes/usecase"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/ioutils"
	middleware "drssr/internal/pkg/middlewares"
	"net/http"

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

	// clothesPrivateAPI.HandleFunc("", ClothesDelivery.getUser).Methods(http.MethodGet)
	clothesPrivateAPI.HandleFunc("", clothesDelivery.addClothes).Methods(http.MethodPost)
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

	// we upload only 1 file
	fileHeader := r.MultipartForm.File["file"][0]
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

	createdClothes, status, err := cd.clothesUseCase.AddFile(ctx, user.ID, fileHeader, file)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to add file: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, createdClothes)
}

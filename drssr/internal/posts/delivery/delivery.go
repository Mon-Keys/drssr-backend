package delivery

import (
	"drssr/internal/models"
	"drssr/internal/pkg/consts"
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/ioutils"
	middleware "drssr/internal/pkg/middlewares"
	"drssr/internal/posts/usecase"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type PostsDelivery struct {
	postsUseCase usecase.IPostUsecase
	logger       logrus.Logger
}

func SetPostsRouting(
	router *mux.Router,
	cu usecase.IPostUsecase,
	authMw middleware.AuthMiddleware,
	logger logrus.Logger,
) {
	postsDelivery := &PostsDelivery{
		postsUseCase: cu,
		logger:       logger,
	}

	// posts API
	looksPrivateAPI := router.PathPrefix("/api/v1/private/posts").Subrouter()
	looksPrivateAPI.Use(middleware.WithRequestID, middleware.WithJSON, authMw.WithAuth)

	looksPrivateAPI.HandleFunc("", postsDelivery.addPost).Methods(http.MethodPost)
	looksPrivateAPI.HandleFunc("", postsDelivery.deletePost).Methods(http.MethodDelete)
	looksPrivateAPI.HandleFunc("/all", postsDelivery.getUserPosts).Methods(http.MethodGet)

	looksPublicAPI := router.PathPrefix("/api/v1/public/posts").Subrouter()
	looksPublicAPI.Use(middleware.WithRequestID, middleware.WithJSON)

	looksPublicAPI.HandleFunc("", postsDelivery.getPost).Methods(http.MethodGet)
	looksPublicAPI.HandleFunc("/all", postsDelivery.getAllPosts).Methods(http.MethodGet)
}

func (pd *PostsDelivery) addPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := pd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	pd.logger = *pd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	var post models.Post
	err := ioutils.ReadJSON(r, &post)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse JSON")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if post.Type == "" || post.ElementID == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Bad request from client")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	post.CreatorID = user.ID

	createdPost, status, err := pd.postsUseCase.AddPost(ctx, user.Email, post)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to save post: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, createdPost)
}

func (pd *PostsDelivery) deletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := pd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	pd.logger = *pd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	pidParam := r.URL.Query().Get("id")
	if pidParam == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid post ID query param")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}
	pid, err := strconv.ParseUint(pidParam, 10, 64)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse post ID")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if pid == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid post ID")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	status, err := pd.postsUseCase.DeletePost(ctx, user.ID, pid)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to delete post: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.SendWithoutBody(w, status)
}

func (pd *PostsDelivery) getUserPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := pd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})
	user := ctx_utils.GetUser(ctx)
	if user == nil {
		logger.WithField("status", http.StatusForbidden).Errorf("Failed to get user from ctx")
		ioutils.SendDefaultError(w, http.StatusForbidden)
		return
	}

	pd.logger = *pd.logger.WithFields(logrus.Fields{
		"user": user.Email,
	}).Logger

	queryParams := r.URL.Query()

	limitStr := queryParams.Get("limit")
	limitInt := 0
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil || limitInt < 0 || limitInt > consts.GetClothesLimit {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse limit: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}

	offsetStr := queryParams.Get("offset")
	offsetInt := 0
	if offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil || offsetInt < 0 {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse offset: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}

	posts, status, err := pd.postsUseCase.GetUserPosts(ctx, user.ID, limitInt, offsetInt)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to get user's posts: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, posts)
}

func (pd *PostsDelivery) getPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := pd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})

	pidParam := r.URL.Query().Get("id")
	if pidParam == "" {
		logger.WithField("status", http.StatusBadRequest).Errorf("Empty look id param")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}
	pid, err := strconv.ParseUint(pidParam, 10, 64)
	if err != nil {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid look id param")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	if pid == 0 {
		logger.WithField("status", http.StatusBadRequest).Errorf("Invalid look id param")
		ioutils.SendDefaultError(w, http.StatusBadRequest)
		return
	}

	post, status, err := pd.postsUseCase.GetPostByID(ctx, pid)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to get post by ID: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, post)
}

func (pd *PostsDelivery) getAllPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	reqID := ctx_utils.GetReqID(ctx)
	logger := pd.logger.WithFields(logrus.Fields{
		"url":    r.URL,
		"req_id": reqID,
	})

	queryParams := r.URL.Query()

	limitStr := queryParams.Get("limit")
	limitInt := 0
	if limitStr != "" {
		limitInt, err := strconv.Atoi(limitStr)
		if err != nil || limitInt < 0 || limitInt > consts.GetClothesLimit {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse limit: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}

	offsetStr := queryParams.Get("offset")
	offsetInt := 0
	if offsetStr != "" {
		offsetInt, err := strconv.Atoi(offsetStr)
		if err != nil || offsetInt < 0 {
			logger.WithField("status", http.StatusBadRequest).Errorf("Failed to parse offset: %w", err)
			ioutils.SendDefaultError(w, http.StatusBadRequest)
			return
		}
	}

	posts, status, err := pd.postsUseCase.GetAllPosts(ctx, limitInt, offsetInt)
	if err != nil || status != http.StatusOK {
		logger.WithField("status", status).Errorf("Failed to get posts: %w", err)
		ioutils.SendDefaultError(w, status)
		return
	}

	ioutils.Send(w, status, posts)
}

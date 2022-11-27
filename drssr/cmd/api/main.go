package main

import (
	"drssr/config"
	clothes_delivery "drssr/internal/clothes/delivery"
	clothes_repository "drssr/internal/clothes/repository"
	clothes_usecase "drssr/internal/clothes/usecase"

	looks_delivery "drssr/internal/looks/delivery"
	looks_repository "drssr/internal/looks/repository"
	looks_usecase "drssr/internal/looks/usecase"

	posts_delivery "drssr/internal/posts/delivery"
	posts_repository "drssr/internal/posts/repository"
	posts_usecase "drssr/internal/posts/usecase"

	"drssr/internal/pkg/classifier"
	"drssr/internal/pkg/cutter"
	middleware "drssr/internal/pkg/middlewares"
	"drssr/internal/pkg/similarity"
	user_delivery "drssr/internal/users/delivery"
	user_repository "drssr/internal/users/repository"
	user_usecase "drssr/internal/users/usecase"
	"net/http"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	config.SetConfig()

	// logger
	logger := logrus.New()

	// classifier client
	cfc, err := classifier.NewRecognizeApiClient(config.Classifier.URL)
	if err != nil {
		logger.Fatal("Failed to connect to recognizeAPI: ", err)
	}

	// repository
	ur := user_repository.NewPostgresqlRepository(config.Postgres, *logger)
	rdr := user_repository.NewRedisRepository(config.Redis, *logger)
	cr := clothes_repository.NewPostgresqlRepository(config.Postgres, *logger)
	lr := looks_repository.NewPostgresqlRepository(config.Postgres, *logger)
	pr := posts_repository.NewPostgresqlRepository(config.Postgres, *logger)

	// router
	router := mux.NewRouter()

	// tg bot
	bot, err := tgbotapi.NewBotAPI(config.TgBotAPIToken.APIToken)
	if err != nil {
		logger.Fatalf("failed to init tg bot: %w", err)
	}

	// usecase
	uu := user_usecase.NewUserUsecase(bot, ur, rdr, *logger)
	cu := clothes_usecase.NewClothesUsecase(
		cr,
		*cutter.New(
			config.Cutter.URL,
			config.Cutter.Timeout,
		),
		cfc,
		*similarity.New(
			config.Similarity.URL,
			config.Similarity.Timeout,
		),
		*logger,
	)
	lu := looks_usecase.NewLooksUsecase(lr, cr, *logger)
	pu := posts_usecase.NewPostsUsecase(pr, cr, lr, *logger)

	// middlewars
	authMw := middleware.NewAuthMiddleware(uu, *logger)

	// delivery
	user_delivery.SetUserRouting(router, uu, authMw, *logger)
	clothes_delivery.SetClothesRouting(router, cu, authMw, *logger)
	looks_delivery.SetLooksRouting(router, lu, authMw, *logger)
	posts_delivery.SetPostsRouting(router, pu, authMw, *logger)

	srv := &http.Server{
		Handler:      router,
		Addr:         config.Drssr.Port,
		WriteTimeout: http.DefaultClient.Timeout,
		ReadTimeout:  http.DefaultClient.Timeout,
	}
	logger.Infof("starting server at %s\n", srv.Addr)

	// logger.Fatal(srv.ListenAndServeTLS("kit-lokle.crt", "kit-lokle.key"))
	logger.Fatal(srv.ListenAndServe())
}

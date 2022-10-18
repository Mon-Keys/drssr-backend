package main

import (
	"drssr/config"
	"drssr/internal/users/delivery"
	"drssr/internal/users/repository"
	"drssr/internal/users/usecase"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	config.SetConfig()

	// logger
	logger := logrus.New()

	// repository
	ur := repository.NewPostgresqlRepository(config.Postgres, *logger)
	rdr := repository.NewRedisRepository(config.Redis, *logger)

	// router
	router := mux.NewRouter()

	// usecase
	uu := usecase.NewUserUsecase(ur, rdr, *logger)

	// middlewars
	// auth := middleware.NewAuthMiddleware(uu, *logger)

	// delivery
	delivery.SetUserRouting(router, uu, *logger)

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

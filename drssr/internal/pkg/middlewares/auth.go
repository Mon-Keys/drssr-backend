package middleware

import (
	"drssr/internal/pkg/ctx_utils"
	"drssr/internal/pkg/ioutils"
	"drssr/internal/users/usecase"
	"net/http"

	"github.com/sirupsen/logrus"
)

type AuthMiddleware struct {
	userUseCase usecase.IUserUsecase
	logger      logrus.Logger
}

func NewAuthMiddleware(uu usecase.IUserUsecase, logger logrus.Logger) AuthMiddleware {
	return AuthMiddleware{
		userUseCase: uu,
		logger:      logger,
	}
}

func (am AuthMiddleware) WithAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := ctx_utils.GetReqID(ctx)

		cookieToken, err := r.Cookie("session-id")
		if err != nil {
			am.logger.Errorf(
				"[%s] [%s] no cookie [status=%d] [error=%s]",
				reqID,
				r.URL,
				http.StatusUnauthorized,
				err,
			)

			ioutils.SendDefaultError(w, http.StatusUnauthorized)
			return
		}

		user, status, err := am.userUseCase.GetUserByCookie(ctx, cookieToken.Value)
		if err != nil || status != http.StatusOK {
			am.logger.Errorf(
				"[%s] [%s] cookie auth failed with [status=%d] [error=%s]",
				reqID,
				r.URL,
				status,
				err,
			)
			ioutils.SendDefaultError(w, status)
			return
		}

		r = r.WithContext(ctx_utils.SetUser(ctx, user))

		h.ServeHTTP(w, r)
	})
}

package ctx_utils

import (
	"context"
	"drssr/internal/models"
)

type reqIDCtxKey struct{}

func SetReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDCtxKey{}, reqID)
}

func GetReqID(ctx context.Context) string {
	reqID, ok := ctx.Value(reqIDCtxKey{}).(string)
	if !ok {
		return ""
	}

	return reqID
}

type userCtxKey struct{}

func SetUser(ctx context.Context, user models.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, &user)
}

func GetUser(ctx context.Context) *models.User {
	user, ok := ctx.Value(userCtxKey{}).(*models.User)
	if !ok {
		return nil
	}

	return user
}

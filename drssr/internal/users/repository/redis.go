package repository

import (
	"context"
	"drssr/config"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type IRedisRepository interface {
	CreateSession(ctx context.Context, sessionID string, email string, expCookieTime time.Duration) error
	DeleteSession(ctx context.Context, cookie string) error
	CheckSession(ctx context.Context, cookie string) (string, error)
	ProlongSession(ctx context.Context, cookie string, expCookieTime time.Duration) error
}

type redisRepository struct {
	client *redis.Client
	logger logrus.Logger
}

func NewRedisRepository(cfg config.RedisConfig, logger logrus.Logger) IRedisRepository {
	return &redisRepository{
		client: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			Password: cfg.Password,
			DB:       cfg.DB,
		}),
		logger: logger,
	}
}

func (rsr *redisRepository) CreateSession(ctx context.Context, sessionID string, email string, expCookieTime time.Duration) error {
	_, err := rsr.client.SetNX(ctx, sessionID, email, expCookieTime*time.Second).Result()
	return err
}

func (rsr *redisRepository) DeleteSession(ctx context.Context, cookie string) error {
	rsr.client.Del(ctx, cookie).Val()
	return nil
}

func (rsr *redisRepository) CheckSession(ctx context.Context, cookie string) (string, error) {
	val, err := rsr.client.Get(ctx, cookie).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (rsr *redisRepository) ProlongSession(ctx context.Context, cookie string, expCookieTime time.Duration) error {
	rsr.client.Expire(ctx, cookie, expCookieTime*time.Second)
	return nil
}

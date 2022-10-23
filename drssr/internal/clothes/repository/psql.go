package repository

import (
	"context"
	"database/sql"
	"drssr/config"
	"drssr/internal/models"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

type IPostgresqlRepository interface {
	AddClothes(ctx context.Context, clothes models.Clothes) (models.Clothes, error)
	DeleteClothes(ctx context.Context, cid uint64) error
	AddClothesUserBind(ctx context.Context, uid uint64, cid uint64) (uint64, error)
}

type postgresqlRepository struct {
	conn   *pgx.ConnPool
	logger logrus.Logger
}

func NewPostgresqlRepository(cfg config.PostgresConfig, logger logrus.Logger) IPostgresqlRepository {
	connStr := fmt.Sprintf(
		"user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		cfg.User,
		cfg.DBName,
		cfg.Password,
		cfg.Host,
		cfg.Port,
	)

	pgxConnectionConfig, err := pgx.ParseConnectionString(connStr)
	if err != nil {
		logger.Fatalf("Invalid config string: %s", err)
	}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     pgxConnectionConfig,
		MaxConnections: 100,
		AfterConnect:   nil,
		AcquireTimeout: 0,
	})
	if err != nil {
		logger.Fatalf("Error %s occurred during connection to database", err)
	}

	return &postgresqlRepository{conn: pool, logger: logger}
}

func (pr *postgresqlRepository) AddClothes(ctx context.Context, clothes models.Clothes) (models.Clothes, error) {
	var createdClothes models.Clothes
	err := pr.conn.QueryRow(
		`INSERT INTO clothes (type, color, img, mask)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			type,
			color,
			img,
			mask,
			brand,
			sex,
			created_at;`,
		clothes.Type,
		clothes.Color,
		clothes.ImgPath,
		clothes.MaskPath,
	).Scan(
		&createdClothes.ID,
		&createdClothes.Type,
		&createdClothes.Color,
		&createdClothes.ImgPath,
		&createdClothes.MaskPath,
		&createdClothes.Brand,
		&createdClothes.Sex,
		&createdClothes.Ctime,
	)

	if err != nil {
		return models.Clothes{}, err
	}
	return createdClothes, nil
}

func (pr *postgresqlRepository) DeleteClothes(ctx context.Context, cid uint64) error {
	var deletedClothesID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM clothes WHERE id = $1
		RETURNING id;`,
		cid,
	).Scan(
		&deletedClothesID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (pr *postgresqlRepository) AddClothesUserBind(ctx context.Context, uid uint64, cid uint64) (uint64, error) {
	var createdBindID uint64
	err := pr.conn.QueryRow(
		`INSERT INTO clothes_users (clothes_id, user_id)
		VALUES ($1, $2)
		RETURNING
			id;`,
		cid,
		uid,
	).Scan(
		&createdBindID,
	)

	if err != nil {
		return 0, err
	}
	return createdBindID, nil
}

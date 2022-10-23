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
	DeleteClothesUserBind(ctx context.Context, bid uint64) error
	GetClothesMaskByTypeAndSex(ctx context.Context, clothesType string, clothesSex string) ([]models.Clothes, error)
	AddSimilarityBind(ctx context.Context, mainCID uint64, secondCID uint64, percent int) (uint64, error)
	DeleteSimilarityBind(ctx context.Context, bid uint64) error
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
		`INSERT INTO clothes (type, color, img, mask, brand, sex)
		VALUES ($1, $2, $3, $4, $5, $6)
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
		clothes.Brand,
		clothes.Sex,
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
		`DELETE FROM clothes WHERE id = $1 RETURNING id;`,
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

func (pr *postgresqlRepository) DeleteClothesUserBind(ctx context.Context, bid uint64) error {
	var deletedBindID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM clothes_users WHERE id = $1 RETURNING id;`,
		bid,
	).Scan(
		&deletedBindID,
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

func (pr *postgresqlRepository) GetClothesMaskByTypeAndSex(
	ctx context.Context,
	clothesType string,
	clothesSex string,
) ([]models.Clothes, error) {
	rows, err := pr.conn.Query(`SELECT id, mask FROM clothes WHERE type = $1 AND sex = $2;`, clothesType, clothesSex)
	if err != nil {
		return []models.Clothes{}, err
	}
	defer rows.Close()

	var respList []models.Clothes
	var row models.Clothes
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.MaskPath,
		)
		if err != nil {
			return []models.Clothes{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.Clothes{}, err
	}

	return respList, nil
}

func (pr *postgresqlRepository) AddSimilarityBind(
	ctx context.Context,
	mainCID uint64,
	secondCID uint64,
	percent int,
) (uint64, error) {
	var createdSimilarityID uint64
	err := pr.conn.QueryRow(
		`INSERT INTO similarity (clothes1_id, clothes2_id, percent)
		VALUES ($1, $2, $3)
		RETURNING
			id;`,
		mainCID,
		secondCID,
		percent,
	).Scan(
		&createdSimilarityID,
	)

	if err != nil {
		return 0, err
	}
	return createdSimilarityID, nil
}

func (pr *postgresqlRepository) DeleteSimilarityBind(ctx context.Context, bid uint64) error {
	var deletedBindID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM similarity WHERE id = $1 RETURNING id;`,
		bid,
	).Scan(
		&deletedBindID,
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

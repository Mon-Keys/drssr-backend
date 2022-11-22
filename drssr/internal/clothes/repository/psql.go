package repository

import (
	"context"
	"drssr/config"
	"drssr/internal/models"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/sirupsen/logrus"
)

type IPostgresqlRepository interface {
	AddClothes(ctx context.Context, clothes models.Clothes) (models.Clothes, error)
	DeleteClothes(ctx context.Context, cid uint64) error
	UpdateClothes(ctx context.Context, newClothesData models.Clothes) (models.Clothes, error)
	GetClothesMaskByTypeAndSex(ctx context.Context, clothesType string, clothesSex string) ([]models.Clothes, error)
	GetAllClothes(ctx context.Context, limit, offset int) ([]models.Clothes, error)
	GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) ([]models.Clothes, error)
	AddSimilarityBind(ctx context.Context, mainCID uint64, secondCID uint64, percent int) (uint64, error)
	DeleteSimilarityBind(ctx context.Context, bid uint64) error
	DeleteSimilarityBindByClothesID(ctx context.Context, cid uint64) error
	GetClothesByID(ctx context.Context, cid uint64) (models.Clothes, error)
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
		`INSERT INTO clothes (type, img, mask, owner_id)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			type,
			name,
			description,
			img,
			mask,
			owner_id,
			created_at;`,
		clothes.Type,
		clothes.ImgPath,
		clothes.MaskPath,
		clothes.OwnerID,
	).Scan(
		&createdClothes.ID,
		&createdClothes.Type,
		&createdClothes.Name,
		&createdClothes.Desc,
		&createdClothes.ImgPath,
		&createdClothes.MaskPath,
		&createdClothes.OwnerID,
		&createdClothes.Ctime,
	)

	if err != nil {
		return models.Clothes{}, err
	}
	return createdClothes, nil
}

func (pr *postgresqlRepository) UpdateClothes(ctx context.Context, newClothesData models.Clothes) (models.Clothes, error) {
	var updatedClothes models.Clothes
	err := pr.conn.QueryRow(
		`UPDATE clothes
		SET (type, name, description, color, brand, sex, link, price, currency) = ($2, $3, $4, $5, $6, $7, $8, $9, $10)
		WHERE id = $1
		RETURNING
			id,
			type,
			name,
			description,
			color,
			img,
			mask,
			brand,
			sex,
			link,
			price,
			currency,
			owner_id,
			created_at;`,
		newClothesData.ID,
		newClothesData.Type,
		newClothesData.Name,
		newClothesData.Desc,
		newClothesData.Color,
		newClothesData.Brand,
		newClothesData.Sex,
		newClothesData.Link,
		newClothesData.Price,
		newClothesData.Currency,
	).Scan(
		&updatedClothes.ID,
		&updatedClothes.Type,
		&updatedClothes.Name,
		&updatedClothes.Desc,
		&updatedClothes.Color,
		&updatedClothes.ImgPath,
		&updatedClothes.MaskPath,
		&updatedClothes.Brand,
		&updatedClothes.Sex,
		&updatedClothes.Link,
		&updatedClothes.Price,
		&updatedClothes.Currency,
		&updatedClothes.OwnerID,
		&updatedClothes.Ctime,
	)

	if err != nil {
		return models.Clothes{}, err
	}
	return updatedClothes, nil
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
		if err == pgx.ErrNoRows {
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
		if err == pgx.ErrNoRows {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (pr *postgresqlRepository) DeleteSimilarityBindByClothesID(ctx context.Context, cid uint64) error {
	var deletedBindID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM similarity WHERE clothes1_id = $1 OR clothes2_id = $1 RETURNING id;`,
		cid,
	).Scan(
		&deletedBindID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (pr *postgresqlRepository) GetAllClothes(ctx context.Context, limit, offset int) ([]models.Clothes, error) {
	query := `SELECT
		id,
		type,
		name,
		description,
		color,
		img,
		mask,
		brand,
		sex,
		link,
		price,
		currency,
		owner_id,
		created_at
	FROM clothes`
	var l string
	if limit > 0 {
		l = fmt.Sprintf(" LIMIT %d", limit)
	}
	var o string
	if offset > 0 {
		o = fmt.Sprintf(" OFFSET %d", offset)
	}
	rows, err := pr.conn.Query(fmt.Sprintf("%s%s%s;", query, l, o))
	if err != nil {
		return []models.Clothes{}, err
	}
	defer rows.Close()

	var respList []models.Clothes
	var row models.Clothes
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.Type,
			&row.Name,
			&row.Desc,
			&row.Color,
			&row.ImgPath,
			&row.MaskPath,
			&row.Brand,
			&row.Sex,
			&row.Link,
			&row.Price,
			&row.Currency,
			&row.OwnerID,
			&row.Ctime,
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

func (pr *postgresqlRepository) GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) ([]models.Clothes, error) {
	query := `SELECT
		id,
		type,
		name,
		description,
		color,
		img,
		mask,
		brand,
		sex,
		link,
		price,
		currency,
		owner_id,
		created_at
	FROM clothes WHERE owner_id = $1`
	var l string
	if limit > 0 {
		l = fmt.Sprintf(" LIMIT %d", limit)
	}
	var o string
	if offset > 0 {
		o = fmt.Sprintf(" OFFSET %d", offset)
	}
	rows, err := pr.conn.Query(fmt.Sprintf("%s%s%s;", query, l, o), uid)
	if err != nil {
		return []models.Clothes{}, err
	}
	defer rows.Close()

	var respList []models.Clothes
	var row models.Clothes
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.Type,
			&row.Name,
			&row.Desc,
			&row.Color,
			&row.ImgPath,
			&row.MaskPath,
			&row.Brand,
			&row.Sex,
			&row.Link,
			&row.Price,
			&row.Currency,
			&row.OwnerID,
			&row.Ctime,
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

func (pr *postgresqlRepository) GetClothesByID(ctx context.Context, cid uint64) (models.Clothes, error) {
	var clothes models.Clothes
	err := pr.conn.QueryRow(
		`SELECT
			id,
			type,
			name,
			description,
			color,
			img,
			mask,
			brand,
			sex,
			link,
			price,
			currency,
			owner_id,
			created_at
		FROM clothes
		WHERE id = $1;`,
		cid,
	).Scan(
		&clothes.ID,
		&clothes.Type,
		&clothes.Name,
		&clothes.Desc,
		&clothes.Color,
		&clothes.ImgPath,
		&clothes.MaskPath,
		&clothes.Brand,
		&clothes.Sex,
		&clothes.Link,
		&clothes.Price,
		&clothes.Currency,
		&clothes.OwnerID,
		&clothes.Ctime,
	)
	if err != nil {
		return models.Clothes{}, err
	}

	return clothes, nil
}

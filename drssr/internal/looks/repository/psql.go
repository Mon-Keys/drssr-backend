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
	AddLook(ctx context.Context, look models.Look) (models.Look, error)
	DeleteLook(ctx context.Context, lid uint64) error
	AddLookClothesBind(ctx context.Context, lid uint64, cid uint64) (uint64, error)
	DeleteLookClothesBind(ctx context.Context, bid uint64) error
	// AddClothesUserBind(ctx context.Context, uid uint64, cid uint64) (uint64, error)
	// DeleteClothesUserBind(ctx context.Context, bid uint64) error
	// GetClothesMaskByTypeAndSex(ctx context.Context, clothesType string, clothesSex string) ([]models.Clothes, error)
	// GetAllClothes(ctx context.Context, limit, offset int) ([]models.Clothes, error)
	// GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) ([]models.Clothes, error)
	// AddSimilarityBind(ctx context.Context, mainCID uint64, secondCID uint64, percent int) (uint64, error)
	// DeleteSimilarityBind(ctx context.Context, bid uint64) error
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

func (pr *postgresqlRepository) AddLook(ctx context.Context, look models.Look) (models.Look, error) {
	var createdLook models.Look
	err := pr.conn.QueryRow(
		`INSERT INTO looks (preview, img, description, creator_id)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			preview,
			img,
			description,
			creator_id,
			created_at;`,
		look.PreviewPath,
		look.ImgPath,
		look.Description,
		look.CreatorID,
	).Scan(
		&createdLook.ID,
		&createdLook.PreviewPath,
		&createdLook.ImgPath,
		&createdLook.Description,
		&createdLook.CreatorID,
		&createdLook.Ctime,
	)

	if err != nil {
		return models.Look{}, err
	}
	return createdLook, nil
}

func (pr *postgresqlRepository) DeleteLook(ctx context.Context, lid uint64) error {
	var deletedLookID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM looks WHERE id = $1 RETURNING id;`,
		lid,
	).Scan(
		&deletedLookID,
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

func (pr *postgresqlRepository) AddLookClothesBind(
	ctx context.Context,
	lid uint64,
	cid uint64,
) (uint64, error) {
	var createdBindID uint64
	err := pr.conn.QueryRow(
		`INSERT INTO clothes_looks (clothes_id, look_id)
		VALUES ($1, $2)
		RETURNING
			id;`,
		cid,
		lid,
	).Scan(
		&createdBindID,
	)

	if err != nil {
		return 0, err
	}
	return createdBindID, nil
}

func (pr *postgresqlRepository) DeleteLookClothesBind(ctx context.Context, bid uint64) error {
	var deletedBindID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM clothes_looks WHERE id = $1 RETURNING id;`,
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

// func (pr *postgresqlRepository) AddClothesUserBind(ctx context.Context, uid uint64, cid uint64) (uint64, error) {
// 	var createdBindID uint64
// 	err := pr.conn.QueryRow(
// 		`INSERT INTO clothes_users (clothes_id, user_id)
// 		VALUES ($1, $2)
// 		RETURNING
// 			id;`,
// 		cid,
// 		uid,
// 	).Scan(
// 		&createdBindID,
// 	)

// 	if err != nil {
// 		return 0, err
// 	}
// 	return createdBindID, nil
// }

// func (pr *postgresqlRepository) DeleteClothesUserBind(ctx context.Context, bid uint64) error {
// 	var deletedBindID uint64
// 	err := pr.conn.QueryRow(
// 		`DELETE FROM clothes_users WHERE id = $1 RETURNING id;`,
// 		bid,
// 	).Scan(
// 		&deletedBindID,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil
// 		} else {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func (pr *postgresqlRepository) GetClothesMaskByTypeAndSex(
// 	ctx context.Context,
// 	clothesType string,
// 	clothesSex string,
// ) ([]models.Clothes, error) {
// 	rows, err := pr.conn.Query(`SELECT id, mask FROM clothes WHERE type = $1 AND sex = $2;`, clothesType, clothesSex)
// 	if err != nil {
// 		return []models.Clothes{}, err
// 	}
// 	defer rows.Close()

// 	var respList []models.Clothes
// 	var row models.Clothes
// 	for rows.Next() {
// 		err := rows.Scan(
// 			&row.ID,
// 			&row.MaskPath,
// 		)
// 		if err != nil {
// 			return []models.Clothes{}, err
// 		}
// 		respList = append(respList, row)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return []models.Clothes{}, err
// 	}

// 	return respList, nil
// }

// func (pr *postgresqlRepository) GetAllClothes(ctx context.Context, limit, offset int) ([]models.Clothes, error) {
// 	query := `SELECT id, type, color, img, mask, brand, sex, created_at FROM clothes`
// 	var l string
// 	if limit > 0 {
// 		l = fmt.Sprintf(" LIMIT %d", limit)
// 	}
// 	var o string
// 	if offset > 0 {
// 		o = fmt.Sprintf(" OFFSET %d", offset)
// 	}
// 	rows, err := pr.conn.Query(fmt.Sprintf("%s%s%s;", query, l, o))
// 	if err != nil {
// 		return []models.Clothes{}, err
// 	}
// 	defer rows.Close()

// 	var respList []models.Clothes
// 	var row models.Clothes
// 	for rows.Next() {
// 		err := rows.Scan(
// 			&row.ID,
// 			&row.Type,
// 			&row.Color,
// 			&row.ImgPath,
// 			&row.MaskPath,
// 			&row.Brand,
// 			&row.Sex,
// 			&row.Ctime,
// 		)
// 		if err != nil {
// 			return []models.Clothes{}, err
// 		}
// 		respList = append(respList, row)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return []models.Clothes{}, err
// 	}

// 	return respList, nil
// }

// func (pr *postgresqlRepository) GetUsersClothes(ctx context.Context, limit, offset int, uid uint64) ([]models.Clothes, error) {
// 	query := `SELECT c.id, c.type, c.color, c.img, c.mask, c.brand, c.sex, c.created_at
// 	FROM clothes c JOIN clothes_users cu on c.id = cu.clothes_id AND cu.user_id=$1`
// 	var l string
// 	if limit > 0 {
// 		l = fmt.Sprintf(" LIMIT %d", limit)
// 	}
// 	var o string
// 	if offset > 0 {
// 		o = fmt.Sprintf(" OFFSET %d", offset)
// 	}
// 	rows, err := pr.conn.Query(fmt.Sprintf("%s%s%s;", query, l, o), uid)
// 	if err != nil {
// 		return []models.Clothes{}, err
// 	}
// 	defer rows.Close()

// 	var respList []models.Clothes
// 	var row models.Clothes
// 	for rows.Next() {
// 		err := rows.Scan(
// 			&row.ID,
// 			&row.Type,
// 			&row.Color,
// 			&row.ImgPath,
// 			&row.MaskPath,
// 			&row.Brand,
// 			&row.Sex,
// 			&row.Ctime,
// 		)
// 		if err != nil {
// 			return []models.Clothes{}, err
// 		}
// 		respList = append(respList, row)
// 	}
// 	if err := rows.Err(); err != nil {
// 		return []models.Clothes{}, err
// 	}

// 	return respList, nil
// }

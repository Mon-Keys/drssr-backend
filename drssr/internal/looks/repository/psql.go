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
	AddLook(ctx context.Context, look models.Look) (models.Look, error)
	DeleteLook(ctx context.Context, lid uint64) (models.Look, error)
	AddLookClothesBind(ctx context.Context, clothes models.ClothesStruct, lid uint64) (models.ClothesStruct, error)
	DeleteLookClothesBind(ctx context.Context, bid uint64) error
	GetLookByID(ctx context.Context, lid uint64) (models.Look, error)
	UpdateLookByID(ctx context.Context, lid uint64, newLook models.Look) (models.Look, error)
	DeleteLookClothesBindsByID(ctx context.Context, lid uint64) ([]models.ClothesStruct, error)
	GetLookClothes(ctx context.Context, lid uint64) ([]models.ClothesStruct, error)
	GetUserLooks(ctx context.Context, limit, offset int, uid uint64) ([]models.Look, error)
	GetAllLooks(ctx context.Context, limit, offset int) ([]models.Look, error)
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
		`INSERT INTO looks (img, name, description, creator_id)
		VALUES ($1, $2, $3, $4)
		RETURNING
			id,
			img,
			name,
			description,
			creator_id,
			created_at;`,
		look.ImgPath,
		look.Name,
		look.Desc,
		look.CreatorID,
	).Scan(
		&createdLook.ID,
		&createdLook.ImgPath,
		&createdLook.Name,
		&createdLook.Desc,
		&createdLook.CreatorID,
		&createdLook.Ctime,
	)

	if err != nil {
		return models.Look{}, err
	}
	return createdLook, nil
}

func (pr *postgresqlRepository) DeleteLook(ctx context.Context, lid uint64) (models.Look, error) {
	var deletedLook models.Look
	err := pr.conn.QueryRow(
		`DELETE FROM looks WHERE id = $1
		RETURNING
			id,
			img,
			name,
			description,
			creator_id,
			created_at;`,
		lid,
	).Scan(
		&deletedLook.ID,
		&deletedLook.ImgPath,
		&deletedLook.Name,
		&deletedLook.Desc,
		&deletedLook.CreatorID,
		&deletedLook.Ctime,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Look{}, nil
		} else {
			return models.Look{}, err
		}
	}
	return deletedLook, nil
}

func (pr *postgresqlRepository) AddLookClothesBind(
	ctx context.Context,
	clothes models.ClothesStruct,
	lid uint64,
) (models.ClothesStruct, error) {
	var createdBind models.ClothesStruct
	err := pr.conn.QueryRow(
		`INSERT INTO clothes_looks (clothes_id, look_id, x, y)
		VALUES ($1, $2, $3, $4)
		RETURNING id, x, y;`,
		clothes.ID,
		lid,
		clothes.Coords.X,
		clothes.Coords.Y,
	).Scan(
		&createdBind.ID,
		&createdBind.Coords.X,
		&createdBind.Coords.Y,
	)

	if err != nil {
		return models.ClothesStruct{}, err
	}
	return createdBind, nil
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
		if err == pgx.ErrNoRows {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (pr *postgresqlRepository) GetLookByID(ctx context.Context, lid uint64) (models.Look, error) {
	var look models.Look
	err := pr.conn.QueryRow(
		`SELECT id, img, name, description, creator_id, created_at FROM looks WHERE id = $1;`,
		lid,
	).Scan(
		&look.ID,
		&look.ImgPath,
		&look.Name,
		&look.Desc,
		&look.CreatorID,
		&look.Ctime,
	)
	if err != nil {
		return models.Look{}, err
	}

	return look, nil
}

func (pr *postgresqlRepository) UpdateLookByID(
	ctx context.Context,
	lid uint64,
	newLook models.Look,
) (models.Look, error) {
	var updatedLook models.Look
	err := pr.conn.QueryRow(
		`UPDATE looks
		SET (img, name, description) = ($2, $3, $4)
		WHERE id = $1
		RETURNING id, img, name, description, creator_id, created_at;`,
		lid,
		newLook.ImgPath,
		newLook.Name,
		newLook.Desc,
	).Scan(
		&updatedLook.ID,
		&updatedLook.ImgPath,
		&updatedLook.Name,
		&updatedLook.Desc,
		&updatedLook.CreatorID,
		&updatedLook.Ctime,
	)

	if err != nil {
		return models.Look{}, err
	}
	return updatedLook, nil
}

func (pr *postgresqlRepository) GetLookClothes(
	ctx context.Context,
	lid uint64,
) ([]models.ClothesStruct, error) {
	rows, err := pr.conn.Query(
		`SELECT
			c.id, c.type, c.name, c.description, c.brand, c.img, c.mask, cl.x, cl.y
		FROM clothes_looks cl
		JOIN clothes c ON c.id = cl.clothes_id
		WHERE look_id = $1;`,
		lid,
	)
	if err != nil {
		return []models.ClothesStruct{}, err
	}
	defer rows.Close()

	var respList []models.ClothesStruct
	var row models.ClothesStruct
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.Type,
			&row.Name,
			&row.Desc,
			&row.Brand,
			&row.ImgPath,
			&row.MaskPath,
			&row.Coords.X,
			&row.Coords.Y,
		)
		if err != nil {
			return []models.ClothesStruct{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.ClothesStruct{}, err
	}

	return respList, nil
}

func (pr *postgresqlRepository) DeleteLookClothesBindsByID(
	ctx context.Context,
	lid uint64,
) ([]models.ClothesStruct, error) {
	rows, err := pr.conn.Query(
		`DELETE FROM clothes_looks WHERE look_id = $1 RETURNING look_id, x, y;`,
		lid,
	)
	if err != nil {
		return []models.ClothesStruct{}, err
	}
	defer rows.Close()

	var respList []models.ClothesStruct
	var row models.ClothesStruct
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.Coords.X,
			&row.Coords.Y,
		)
		if err != nil {
			return []models.ClothesStruct{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.ClothesStruct{}, err
	}

	return respList, nil
}

func (pr *postgresqlRepository) GetUserLooks(ctx context.Context, limit, offset int, uid uint64) ([]models.Look, error) {
	query := `SELECT
		id,
		img,
		name,
		description,
		creator_id,
		created_at
	FROM looks WHERE creator_id = $1`
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
		return []models.Look{}, err
	}
	defer rows.Close()

	var respList []models.Look
	var row models.Look
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.ImgPath,
			&row.Name,
			&row.Desc,
			&row.CreatorID,
			&row.Ctime,
		)
		if err != nil {
			return []models.Look{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.Look{}, err
	}

	return respList, nil
}

func (pr *postgresqlRepository) GetAllLooks(ctx context.Context, limit, offset int) ([]models.Look, error) {
	query := `SELECT
		id,
		img,
		name,
		description,
		creator_id,
		created_at
	FROM looks`
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
		return []models.Look{}, err
	}
	defer rows.Close()

	var respList []models.Look
	var row models.Look
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.ImgPath,
			&row.Name,
			&row.Desc,
			&row.CreatorID,
			&row.Ctime,
		)
		if err != nil {
			return []models.Look{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.Look{}, err
	}

	return respList, nil
}

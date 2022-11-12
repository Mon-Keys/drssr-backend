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
	AddPost(ctx context.Context, post models.Post) (models.Post, error)
	DeletePost(ctx context.Context, pid uint64) (models.Post, error)

	GetPostByID(ctx context.Context, pid uint64) (models.Post, error)
	GetUserPosts(ctx context.Context, limit, offset int, uid uint64) ([]models.Post, error)
	GetAllPosts(ctx context.Context, limit, offset int) ([]models.Post, error)

	// AddLookClothesBind(ctx context.Context, clothes models.ClothesStruct, lid uint64) (models.ClothesStruct, error)
	// DeleteLookClothesBind(ctx context.Context, bid uint64) error
	// GetLookByID(ctx context.Context, lid uint64) (models.Look, error)
	// UpdateLookByID(ctx context.Context, lid uint64, newLook models.Look) (models.Look, error)
	// DeleteLookClothesBindsByID(ctx context.Context, lid uint64) ([]models.ClothesStruct, error)
	// GetLookClothes(ctx context.Context, lid uint64) ([]models.ClothesStruct, error)
	// GetUserLooks(ctx context.Context, limit, offset int, uid uint64) ([]models.Look, error)
	// GetAllLooks(ctx context.Context, limit, offset int) ([]models.Look, error)
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

func (pr *postgresqlRepository) AddPost(ctx context.Context, post models.Post) (models.Post, error) {
	var createdPost models.Post
	err := pr.conn.QueryRow(
		`INSERT INTO posts (type, description, element_id, creator_id, previews_paths)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			type,
			description,
			element_id,
			creator_id,
			previews_paths,
			created_at;`,
		post.Type,
		post.Description,
		post.ElementID,
		post.CreatorID,
		post.PreviewsPaths,
	).Scan(
		&createdPost.ID,
		&createdPost.Type,
		&createdPost.Description,
		&createdPost.ElementID,
		&createdPost.CreatorID,
		&createdPost.PreviewsPaths,
		&createdPost.Ctime,
	)

	if err != nil {
		return models.Post{}, err
	}
	return createdPost, nil
}

func (pr *postgresqlRepository) DeletePost(ctx context.Context, pid uint64) (models.Post, error) {
	var deletedPost models.Post
	err := pr.conn.QueryRow(
		`DELETE FROM posts WHERE id = $1
		RETURNING
			id,
			type,
			description,
			element_id,
			creator_id,
			previews_paths,
			created_at;`,
		pid,
	).Scan(
		&deletedPost.ID,
		&deletedPost.Type,
		&deletedPost.Description,
		&deletedPost.ElementID,
		&deletedPost.CreatorID,
		&deletedPost.PreviewsPaths,
		&deletedPost.Ctime,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return models.Post{}, nil
		} else {
			return models.Post{}, err
		}
	}
	return deletedPost, nil
}

func (pr *postgresqlRepository) GetPostByID(ctx context.Context, pid uint64) (models.Post, error) {
	var post models.Post
	err := pr.conn.QueryRow(
		`SELECT id, type, description, element_id, creator_id, previews_paths, created_at FROM posts WHERE id = $1;`,
		pid,
	).Scan(
		&post.ID,
		&post.Type,
		&post.Description,
		&post.ElementID,
		&post.CreatorID,
		&post.PreviewsPaths,
		&post.Ctime,
	)
	if err != nil {
		return models.Post{}, err
	}

	return post, nil
}

func (pr *postgresqlRepository) GetUserPosts(ctx context.Context, limit, offset int, uid uint64) ([]models.Post, error) {
	query := `SELECT
		id,
		type,
		description,
		element_id,
		creator_id,
		previews_paths,
		created_at
	FROM posts WHERE creator_id = $1;`
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
		return []models.Post{}, err
	}
	defer rows.Close()

	var respList []models.Post
	var row models.Post
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.Type,
			&row.Description,
			&row.ElementID,
			&row.CreatorID,
			&row.PreviewsPaths,
			&row.Ctime,
		)
		if err != nil {
			return []models.Post{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.Post{}, err
	}

	return respList, nil
}

func (pr *postgresqlRepository) GetAllPosts(ctx context.Context, limit, offset int) ([]models.Post, error) {
	query := `SELECT
		id,
		type,
		description,
		element_id,
		creator_id,
		previews_paths,
		created_at
	FROM posts;`
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
		return []models.Post{}, err
	}
	defer rows.Close()

	var respList []models.Post
	var row models.Post
	for rows.Next() {
		err := rows.Scan(
			&row.ID,
			&row.Type,
			&row.Description,
			&row.ElementID,
			&row.CreatorID,
			&row.PreviewsPaths,
			&row.Ctime,
		)
		if err != nil {
			return []models.Post{}, err
		}
		respList = append(respList, row)
	}
	if err := rows.Err(); err != nil {
		return []models.Post{}, err
	}

	return respList, nil
}

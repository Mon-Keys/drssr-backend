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
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByNickname(ctx context.Context, nickname string) (models.User, error)
	AddUser(ctx context.Context, user models.SignupCredentials) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, user models.User) (models.User, error)
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

func (pr *postgresqlRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User
	err := pr.conn.QueryRow(
		`SELECT
			id,
			nickname,
			email,
			password,
			name,
			avatar,
			stylist,
			birth_date,
			description,
			created_at
		FROM
			users
		WHERE
			email = $1;`,
		email,
	).Scan(
		&user.ID,
		&user.Nickname,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Avatar,
		&user.Stylist,
		&user.BirthDate,
		&user.Desc,
		&user.Ctime,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (pr *postgresqlRepository) GetUserByNickname(ctx context.Context, nickname string) (models.User, error) {
	var user models.User
	err := pr.conn.QueryRow(
		`SELECT
			id,
			nickname,
			email,
			password,
			name,
			avatar,
			stylist,
			birth_date,
			description,
			created_at
		FROM
			users
		WHERE
			nickname = $1;`,
		nickname,
	).Scan(
		&user.ID,
		&user.Nickname,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Avatar,
		&user.Stylist,
		&user.BirthDate,
		&user.Desc,
		&user.Ctime,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (pr *postgresqlRepository) AddUser(ctx context.Context, user models.SignupCredentials) (models.User, error) {
	var createdUser models.User
	err := pr.conn.QueryRow(
		`INSERT INTO users (nickname, email, password, name, birth_date, description)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, nickname, email, name, birth_date, description, created_at;`,
		user.Nickname,
		user.Email,
		user.Password,
		user.Name,
		user.BirthDate,
		user.Desc,
	).Scan(
		&createdUser.ID,
		&createdUser.Nickname,
		&createdUser.Email,
		&createdUser.Name,
		&createdUser.BirthDate,
		&createdUser.Desc,
		&createdUser.Ctime,
	)

	if err != nil {
		return models.User{}, err
	}
	return createdUser, nil
}

func (pr *postgresqlRepository) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
}

func (pr *postgresqlRepository) DeleteUser(ctx context.Context, user models.User) (models.User, error) {
}

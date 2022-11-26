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
	GetUserByEmailOrNickname(ctx context.Context, email string, nickname string) (models.User, error)
	GetUserByLogin(ctx context.Context, login string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByNickname(ctx context.Context, nickname string) (models.User, error)
	AddUser(ctx context.Context, user models.SignupCredentials) (models.User, error)
	UpdateUser(ctx context.Context, user models.UpdateUserReq) (models.User, error)
	DeleteUser(ctx context.Context, uid uint64) error
	UpdateAvatar(ctx context.Context, uid uint64, newAvatar string) (models.User, error)

	CheckStatus(ctx context.Context) (int, error)

	BecomeStylist(ctx context.Context, uid uint64) (models.User, error)
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

func (pr *postgresqlRepository) CheckStatus(ctx context.Context) (int, error) {
	var counter int
	err := pr.conn.QueryRow(
		`SELECT COUNT(*) FROM users;`,
	).Scan(
		&counter,
	)
	if err != nil {
		return 0, err
	}

	return counter, nil
}

func (pr *postgresqlRepository) GetUserByEmailOrNickname(
	ctx context.Context,
	email string,
	nickname string,
) (models.User, error) {
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
			date_part('year', age(birth_date)) as age,
			description,
			created_at
		FROM
			users
		WHERE
			email = $1 OR nickname = $2;`,
		email,
		nickname,
	).Scan(
		&user.ID,
		&user.Nickname,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Avatar,
		&user.Stylist,
		&user.Age,
		&user.Desc,
		&user.Ctime,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (pr *postgresqlRepository) GetUserByLogin(
	ctx context.Context,
	login string,
) (models.User, error) {
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
			date_part('year', age(birth_date)) as age,
			description,
			created_at
		FROM
			users
		WHERE
			email = $1 OR nickname = $1;`,
		login,
	).Scan(
		&user.ID,
		&user.Nickname,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Avatar,
		&user.Stylist,
		&user.Age,
		&user.Desc,
		&user.Ctime,
	)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
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
			date_part('year', age(birth_date)) as age,
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
		&user.Age,
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
			date_part('year', age(birth_date)) as age,
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
		&user.Age,
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
		RETURNING
			id,
			nickname,
			email,
			password,
			name,
			avatar,
			stylist,
			date_part('year', age(birth_date)) as age,
			description,
			created_at;`,
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
		&createdUser.Password,
		&createdUser.Name,
		&createdUser.Avatar,
		&createdUser.Stylist,
		&createdUser.Age,
		&createdUser.Desc,
		&createdUser.Ctime,
	)

	if err != nil {
		return models.User{}, err
	}
	return createdUser, nil
}

func (pr *postgresqlRepository) UpdateUser(ctx context.Context, newUserData models.UpdateUserReq) (models.User, error) {
	var updatedUser models.User
	err := pr.conn.QueryRow(
		`UPDATE users
		SET (nickname, name, birth_date, description) = ($2, $3, $4, $5, $6)
		WHERE email = $1
		RETURNING
			id,
			nickname,
			email,
			password,
			name,
			avatar,
			stylist,
			date_part('year', age(birth_date)) as age,
			description,
			created_at;`,
		newUserData.Email,
		newUserData.Nickname,
		newUserData.Name,
		newUserData.BirthDate,
		newUserData.Desc,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Nickname,
		&updatedUser.Email,
		&updatedUser.Password,
		&updatedUser.Name,
		&updatedUser.Avatar,
		&updatedUser.Stylist,
		&updatedUser.Age,
		&updatedUser.Desc,
		&updatedUser.Ctime,
	)

	if err != nil {
		return models.User{}, err
	}
	return updatedUser, nil
}

func (pr *postgresqlRepository) DeleteUser(ctx context.Context, uid uint64) error {
	var deletedUserID uint64
	err := pr.conn.QueryRow(
		`DELETE FROM users WHERE id = $1
		RETURNING id;`,
		uid,
	).Scan(
		&deletedUserID,
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

func (pr *postgresqlRepository) BecomeStylist(ctx context.Context, uid uint64) (models.User, error) {
	var updatedUser models.User
	err := pr.conn.QueryRow(
		`UPDATE users
		SET stylist = true
		WHERE id = $1
		RETURNING
			id,
			nickname,
			email,
			password,
			name,
			avatar,
			stylist,
			date_part('year', age(birth_date)) as age,
			description,
			created_at;`,
		uid,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Nickname,
		&updatedUser.Email,
		&updatedUser.Password,
		&updatedUser.Name,
		&updatedUser.Avatar,
		&updatedUser.Stylist,
		&updatedUser.Age,
		&updatedUser.Desc,
		&updatedUser.Ctime,
	)

	if err != nil {
		return models.User{}, err
	}
	return updatedUser, nil
}

func (pr *postgresqlRepository) UpdateAvatar(ctx context.Context, uid uint64, newAvatar string) (models.User, error) {
	var updatedUser models.User
	err := pr.conn.QueryRow(
		`UPDATE users
		SET avatar = $2
		WHERE id = $1
		RETURNING
			id,
			nickname,
			email,
			password,
			name,
			avatar,
			stylist,
			date_part('year', age(birth_date)) as age,
			description,
			created_at;`,
		uid,
		newAvatar,
	).Scan(
		&updatedUser.ID,
		&updatedUser.Nickname,
		&updatedUser.Email,
		&updatedUser.Password,
		&updatedUser.Name,
		&updatedUser.Avatar,
		&updatedUser.Stylist,
		&updatedUser.Age,
		&updatedUser.Desc,
		&updatedUser.Ctime,
	)

	if err != nil {
		return models.User{}, err
	}
	return updatedUser, nil
}

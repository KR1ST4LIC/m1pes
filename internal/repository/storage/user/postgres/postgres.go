package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx"

	"m1pes/internal/config"
	"m1pes/internal/models"
)

type Repository struct {
	Conn *pgx.ConnPool
}

func New(cfg config.DBConnConfig) *Repository {
	conn, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig: pgx.ConnConfig{
			Host:     cfg.Host,
			Port:     uint16(cfg.Port),
			User:     cfg.Username,
			Password: cfg.Password,
			Database: cfg.Database,
		},
	})
	if err != nil {
		panic(err)
	}

	return &Repository{Conn: conn}
}

func (r *Repository) ChangeBalance(ctx context.Context, userId, amount int64) error {
	cmd, err := r.Conn.ExecEx(ctx, "UPDATE users SET bal=bal+$1 WHERE tg_id=$2;", nil, amount, userId)
	if err != nil {
		return err
	}
	fmt.Println(cmd.RowsAffected())
	return nil
}

func (r *Repository) NewUser(ctx context.Context, user models.User) error {
	_, err := r.Conn.ExecEx(ctx, "INSERT INTO users(tg_id, bal, capital, percent, income) VALUES($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING;", nil, user.Id, user.Balance, user.Capital, 0.01, user.Income)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetUser(ctx context.Context, userId int64) (models.User, error) {
	var user models.User
	res := r.Conn.QueryRowEx(ctx, "SELECT bal, capital, percent, income FROM users WHERE tg_id=$1", nil, userId)
	err := res.Scan(&user.Balance, &user.Capital, &user.Percent, &user.Income)
	if err != nil {
		return models.User{}, err
	}
	user.Id = userId
	return user, nil
}

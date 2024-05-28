package postgres

import (
	"context"

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

func (r *Repository) ChangeBalance(ctx context.Context, userId int64, amount float64) error {
	_, err := r.Conn.ExecEx(ctx, "UPDATE users SET bal=bal+$1 WHERE tg_id=$2;", nil, amount, userId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) NewUser(ctx context.Context, user models.User) error {
	_, err := r.Conn.ExecEx(ctx, "INSERT INTO users(tg_id, percent) VALUES($1, $2) ON CONFLICT DO NOTHING;", nil, user.Id, 0.01)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetUser(ctx context.Context, userId int64) (models.User, error) {
	var user models.User
	res := r.Conn.QueryRowEx(ctx, "SELECT bal, capital, percent FROM users WHERE tg_id=$1", nil, userId)
	err := res.Scan(&user.Balance, &user.Capital, &user.Percent)
	if err != nil {
		return models.User{}, err
	}
	user.Id = userId
	return user, nil
}

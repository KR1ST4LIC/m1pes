package postgres

import (
	"context"
	"github.com/jackc/pgx"
	"m1pes/internal/config"
	"m1pes/internal/models"
)

type Repository struct {
	Conn *pgx.Conn
}

func New(cfg config.DBConnConfig) *Repository {
	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     cfg.Host,
		Port:     uint16(cfg.Port),
		User:     cfg.Username,
		Password: cfg.Password,
		Database: cfg.Database,
	})
	if err != nil {
		panic(err)
	}

	return &Repository{Conn: conn}
}

func (r *Repository) NewUser(ctx context.Context, user models.User) error {
	_, err := r.Conn.ExecEx(ctx, "INSERT INTO users(tg_id) VALUES($1) ON CONFLICT DO NOTHING;", nil, user.Id)
	if err != nil {
		return err
	}
	return nil
}

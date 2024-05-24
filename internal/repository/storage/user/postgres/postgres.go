package postgres

import (
	"github.com/jackc/pgx"
	"m1pes/internal/models"
)

type Repository struct {
	conn *pgx.Conn
}

func New() *Repository {
	conn, err := pgx.Connect(pgx.ConnConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "m1pes-user",
		Password: "m1pepass",
		Database: "m1pes",
	})
	if err != nil {
		panic(err)
	}

	return &Repository{conn: conn}
}

func (r *Repository) NewUser(user models.User) error {
	_, err := r.conn.Exec("INSERT INTO users(tg_id) VALUES($1) ON CONFLICT DO NOTHING;", user.Id)
	if err != nil {
		return err
	}
	return nil
}

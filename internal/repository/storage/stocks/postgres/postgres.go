package postgres

import (
	"github.com/jackc/pgx"
	"m1pes/internal/config"
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

func (r *Repository) GetCoinList(userId int64) ([]string, error) {
	var coinList []string
	rows, err := r.Conn.Query("SELECT coin_name FROM coin WHERE user_id=$1", userId)
	if err != nil {
		return nil, err
	}

	var coin string
	for rows.Next() {
		if err = rows.Scan(&coin); err != nil {
			return nil, err
		}
		coinList = append(coinList, coin)
	}
	return coinList, nil
}

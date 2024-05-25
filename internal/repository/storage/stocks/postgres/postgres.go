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

func (r *Repository) AddCoin(userId int64, coinTag string) error {
	_, err := r.Conn.Exec("INSERT INTO coin (user_id, coin_name) VALUES ($1, $2)", userId, coinTag)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) CheckStatus(userId int64) (string, error) {
	var status string
	rows, err := r.Conn.Query("SELECT status FROM users WHERE tg_id=$1", userId)
	if err != nil {
		return "", err
	}
	var st string
	for rows.Next() {
		if err = rows.Scan(&st); err != nil {
			return "", err
		}
		status = st
	}

	return status, nil
}

func (r *Repository) UpdateStatus(userID int64, status string) error {
	_, err := r.Conn.Exec("UPDATE users SET status = $1 WHERE tg_id=$2", status, userID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) UpdatePercent(userID int64, percent float64) error {
	_, err := r.Conn.Exec("UPDATE users SET percent = $1 WHERE tg_id=$2", percent, userID)
	if err != nil {
		return err
	}

	return nil
}

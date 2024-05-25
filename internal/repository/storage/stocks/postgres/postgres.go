package postgres

import (
	"context"

	"github.com/jackc/pgx"

	"m1pes/internal/models"

	"m1pes/internal/config"
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

func (r *Repository) GetCoin(ctx context.Context, userId int64, coinName string) (models.Coin, error) {
	var coin models.Coin
	rows := r.Conn.QueryRowEx(ctx, "SELECT coin_name, entry_price, decrement, count, buy FROM coin WHERE user_id=$1 AND coin_name=$2", nil, userId, coinName)
	err := rows.Scan(&coin.Name, &coin.EntryPrice, &coin.Decrement, &coin.Count, &coin.Buy)
	if err != nil {
		return coin, err
	}
	return coin, nil
}

func (r *Repository) GetCoinList(ctx context.Context, userId int64) ([]string, error) {
	var coinList []string
	rows, err := r.Conn.QueryEx(ctx, "SELECT coin_name FROM coin WHERE user_id=$1", nil, userId)
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

func (r *Repository) AddCoin(coin models.Coin) error {
	_, err := r.Conn.Exec("INSERT INTO coin (user_id, coin_name, entry_price, decrement, count, buy) VALUES ($1, $2, $3, $4, $5, $6)",
		coin.UserId,
		coin.Name,
		coin.EntryPrice,
		coin.Decrement,
		coin.Count,
		coin.Buy,
	)
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

func (r *Repository) UpdateCoin(userID int64, coinTag string, entryPrice float64) error {
	_, err := r.Conn.Exec("UPDATE coin SET entry_price = $1 WHERE (tg_id,coin_name)=($2,$3)", entryPrice, userID, coinTag)
	if err != nil {
		return err
	}

	return nil
}

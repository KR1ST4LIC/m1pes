package postgres

import "github.com/jackc/pgx"

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

func (r *Repository) GetCoinList(userId int64) ([]string, error) {
	var coinList []string
	rows, err := r.conn.Query("SELECT coin_name FROM coin WHERE user_id=$1", userId)
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

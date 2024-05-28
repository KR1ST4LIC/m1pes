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

func (r *Repository) DeleteCoin(ctx context.Context, userId int64, coinTag string) error {
	_, err := r.Conn.ExecEx(ctx, "DELETE FROM coin WHERE user_id=$1 AND coin_name=$2", nil, userId, coinTag)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) GetCoin(ctx context.Context, userId int64, coinName string) (models.Coin, error) {
	var coin models.Coin
	rows := r.Conn.QueryRowEx(ctx, "SELECT coin_name, entry_price, decrement, count, buy FROM coin WHERE user_id=$1 AND coin_name=$2", nil, userId, coinName)
	err := rows.Scan(&coin.Name, &coin.EntryPrice, &coin.Decrement, &coin.Count, &coin.Buy)
	if err != nil {
		return coin, err
	}
	coin.UserId = userId
	return coin, nil
}

func (r *Repository) GetCoinList(ctx context.Context, userId int64) (models.List, error) {
	coinLists := models.List{}
	var coinList []string
	buyList := make(map[string][]float64)
	var countList []float64
	rows, err := r.Conn.QueryEx(ctx, "SELECT coin_name, count, buy FROM coin WHERE user_id=$1;", nil, userId)
	if err != nil {
		return coinLists, err
	}

	var coin string
	var buy []float64
	var count float64
	for rows.Next() {
		if err = rows.Scan(&coin, &count, &buy); err != nil {
			return coinLists, err
		}
		coinList = append(coinList, coin)
		buyList[coin] = buy
		countList = append(countList, count)
	}
	coinLists = models.List{
		Name:  coinList,
		Buy:   buyList,
		Count: countList,
	}
	return coinLists, nil
}

func (r *Repository) AddCoin(coin models.Coin) error {
	_, err := r.Conn.Exec("INSERT INTO coin (user_id, coin_name) VALUES ($1, $2) ON CONFLICT DO NOTHING;",
		coin.UserId,
		coin.Name,
	)
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

func (r *Repository) UpdateCoin(userID int64, coinTag string, entryPrice, percent float64) error {
	_, err := r.Conn.Exec("UPDATE coin SET (entry_price, decrement) = ($1,$4) WHERE (user_id,coin_name)=($2,$3)", entryPrice, userID, coinTag, percent*entryPrice)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateCount(userID int64, count float64, coinTag string, decrement float64, buy []float64) error {
	_, err := r.Conn.Exec("UPDATE coin SET (decrement, count,buy) = ($1,$2,$3) WHERE (user_id,coin_name)=($4,$5)", decrement, count, buy, userID, coinTag)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) SellCoin(userID int64, coinTag string, currentPrice, decrement float64) error {
	_, err := r.Conn.Exec("UPDATE coin SET (decrement, count,buy,entry_price) = ($1,$2,$3,$4) WHERE (user_id,coin_name)=($5,$6)", decrement, 0, nil, currentPrice, userID, coinTag)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) InsertIncome(userID int64, coinTag string, income, count float64) error {
	_, err := r.Conn.Exec("INSERT INTO income (user_id, coin_name,income,count) VALUES ($1, $2, $3, $4) ON CONFLICT DO NOTHING;", userID, coinTag, income, count)
	if err != nil {
		return err
	}
	return nil
}

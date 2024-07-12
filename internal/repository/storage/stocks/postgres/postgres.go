package postgres

import (
	"context"
	"fmt"
	"strings"

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
	rows := r.Conn.QueryRowEx(ctx, "SELECT coin_name, entry_price, decrement, count, buy, buy_order_id, sell_order_id FROM coin WHERE user_id=$1 AND coin_name=$2;", nil, userId, coinName)
	err := rows.Scan(&coin.Name, &coin.EntryPrice, &coin.Decrement, &coin.Count, &coin.Buy, &coin.BuyOrderId, &coin.SellOrderId)
	if err != nil {
		return coin, err
	}
	coin.UserId = userId
	return coin, nil
}

func (r *Repository) GetCoiniks(ctx context.Context, coinName string) (models.Coiniks, error) {
	var coiniks models.Coiniks
	rows := r.Conn.QueryRowEx(ctx, "SELECT qty_decimals, price_decimals, min_sum_buy FROM coiniks WHERE coin_name=$1;", nil, coinName)
	err := rows.Scan(&coiniks.QtyDecimals, &coiniks.PriceDecimals, &coiniks.MinSumBuy)
	if err != nil {
		return coiniks, err
	}
	coiniks.Name = coinName
	return coiniks, nil
}

func (r *Repository) EditBuy(ctx context.Context, userId int64, buy bool) error {
	_, err := r.Conn.Exec("UPDATE users SET buy = $1 WHERE tg_id = $2;", buy, userId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) ExistCoin(ctx context.Context, coinTag string) (bool, error) {
	var exist bool
	rows := r.Conn.QueryRowEx(ctx, "SELECT EXISTS(SELECT coin_name FROM coiniks WHERE coin_name = $1);", nil, coinTag)
	err := rows.Scan(&exist)
	if err != nil {
		return false, err
	}
	return exist, nil
}

func (r *Repository) GetCoinList(ctx context.Context, userId int64) ([]models.Coin, error) {
	coinList := make([]models.Coin, 0)
	rows, err := r.Conn.QueryEx(ctx, "SELECT coin_name, count, buy, entry_price, user_id, decrement, buy_order_id, sell_order_id FROM coin WHERE user_id=$1;", nil, userId)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		coin := models.Coin{}
		if err = rows.Scan(&coin.Name, &coin.Count, &coin.Buy, &coin.EntryPrice, &coin.UserId, &coin.Decrement, &coin.BuyOrderId, &coin.SellOrderId); err != nil {
			return nil, err
		}
		coinList = append(coinList, coin)
	}

	return coinList, nil
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

func generateUpdateCoinQuery(coin models.Coin) (string, []interface{}, error) {
	if coin.UserId == 0 {
		return "", nil, fmt.Errorf("user ID is required")
	}
	if coin.Name == "" {
		return "", nil, fmt.Errorf("name is required")
	}

	tableName := "coin"
	var setClauses []string
	var values []interface{}
	i := 1

	if coin.EntryPrice != 0 {
		setClauses = append(setClauses, fmt.Sprintf("entry_price = $%d", i))
		values = append(values, coin.EntryPrice)
		i++
	}
	if coin.Count != 0 {
		setClauses = append(setClauses, fmt.Sprintf("count = $%d", i))
		values = append(values, coin.Count)
		i++
	}
	if coin.Buy != nil {
		setClauses = append(setClauses, fmt.Sprintf("buy = $%d", i))
		values = append(values, coin.Buy)
		i++
	}
	if coin.Decrement != 0 {
		setClauses = append(setClauses, fmt.Sprintf("decrement = $%d", i))
		values = append(values, coin.Decrement)
		i++
	}
	if coin.BuyOrderId != "" {
		setClauses = append(setClauses, fmt.Sprintf("buy_order_id = $%d", i))
		values = append(values, coin.BuyOrderId)
		i++
	}
	if coin.SellOrderId != "" {
		setClauses = append(setClauses, fmt.Sprintf("sell_order_id = $%d", i))
		if coin.SellOrderId == "setNull" {
			values = append(values, "")
		} else {
			values = append(values, coin.SellOrderId)
		}
		i++
	}

	if len(setClauses) == 0 {
		return "", nil, fmt.Errorf("from generateUpdateCoinQuery: no fields to update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE user_id = $%d AND coin_name=$%d", tableName, strings.Join(setClauses, ", "), i, i+1)
	values = append(values, coin.UserId, coin.Name)

	return query, values, nil
}

func (r *Repository) UpdateCoin(ctx context.Context, coin models.Coin) error {
	query, values, err := generateUpdateCoinQuery(coin)
	if err != nil {
		return err
	}

	_, err = r.Conn.ExecEx(ctx, query, nil, values...)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateCount(userID int64, count float64, coinTag string, decrement float64, buy []float64) error {
	_, err := r.Conn.Exec("UPDATE coin SET (decrement, count,buy) = ($1,$2,$3) WHERE (user_id,coin_name)=($4,$5);", decrement, count, buy, userID, coinTag)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) ResetCoin(ctx context.Context, coin models.Coin, user models.User) error {
	_, err := r.Conn.ExecEx(ctx, "UPDATE coin SET (entry_price, decrement) = ($1,$4) WHERE (user_id,coin_name)=($2,$3);", nil, coin.EntryPrice, user.Id, coin.Name, user.Percent*coin.EntryPrice)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) SellCoin(userID int64, coinTag string, sellPrice float64) error {
	_, err := r.Conn.Exec("UPDATE coin SET (count,buy,entry_price) = ($1,$2,$3) WHERE (user_id,coin_name)=($4,$5);", 0, nil, sellPrice, userID, coinTag)
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

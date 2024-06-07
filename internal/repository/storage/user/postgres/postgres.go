package postgres

import (
	"context"
	"fmt"
	"strings"

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

func generateUpdateUserQuery(user models.User) (string, []interface{}, error) {
	if user.Id == 0 {
		return "", nil, fmt.Errorf("user ID is required")
	}

	tableName := "users"
	var setClauses []string
	var values []interface{}
	i := 1

	if user.Percent != 0 {
		setClauses = append(setClauses, fmt.Sprintf("percent = $%d", i))
		values = append(values, user.Percent)
		i++
	}
	if user.Balance != 0 {
		setClauses = append(setClauses, fmt.Sprintf("bal = $%d", i))
		values = append(values, user.Balance)
		i++
	}
	if user.Capital != 0 {
		setClauses = append(setClauses, fmt.Sprintf("capital = $%d", i))
		values = append(values, user.Capital)
		i++
	}
	if user.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", i))
		values = append(values, user.Status)
		i++
	}
	setClauses = append(setClauses, fmt.Sprintf("trading_activated = $%d", i))
	values = append(values, user.TradingActivated)
	i++

	if len(setClauses) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE tg_id = $%d", tableName, strings.Join(setClauses, ", "), i)
	values = append(values, user.Id)

	return query, values, nil
}

func (r *Repository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	rows, err := r.Conn.QueryEx(ctx, "SELECT tg_id, bal, capital, percent, status, trading_activated FROM users", nil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]models.User, 0)
	for rows.Next() {
		user := models.User{}
		err = rows.Scan(&user.Id, &user.Balance, &user.Capital, &user.Percent, &user.Status, &user.TradingActivated)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user models.User) error {
	query, values, err := generateUpdateUserQuery(user)
	if err != nil {
		return err
	}

	_, err = r.Conn.ExecEx(ctx, query, nil, values...)
	if err != nil {
		return err
	}
	return nil
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
	res := r.Conn.QueryRowEx(ctx, "SELECT bal, capital, percent, status, api_key, secret_key FROM users WHERE tg_id=$1;", nil, userId)
	err := res.Scan(&user.Balance, &user.Capital, &user.Percent, &user.Status, &user.ApiKey, &user.SecretKey)
	if err != nil {
		return models.User{}, err
	}
	user.Id = userId
	return user, nil
}

func (r *Repository) GetIncome(ctx context.Context, userId int64) (float64, error) {
	rows, err := r.Conn.QueryEx(ctx, "SELECT income from income where time >= CURRENT_DATE AND user_id = $1;", nil, userId)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	incomes := 0.0
	for rows.Next() {
		income := 0.0
		err = rows.Scan(&income)
		if err != nil {
			return 0, err
		}
		incomes += income
	}
	return incomes, nil
}

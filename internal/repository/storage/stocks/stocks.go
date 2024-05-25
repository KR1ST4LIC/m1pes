package stocks

type Repository interface {
	GetCoinList(userId int64) ([]string, error)
	AddCoin(userId int64, coinTag string) error
	CheckStatus(userId int64) (string, error)
	UpdateStatus(userID int64, status string) error
	UpdatePercent(userID int64, percent float64) error
}

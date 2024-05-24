package stocks

type Repository interface {
	GetCoinList(userId int64) ([]string, error)
}

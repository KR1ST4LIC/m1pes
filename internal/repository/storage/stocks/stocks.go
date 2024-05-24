package stocks

type Repository interface {
	GetCoinList() ([]string, error)
}

package stocks

type Repository interface {
	GetPrice(coinTag string) (float64, error)
}

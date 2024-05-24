package bybit

type Repository struct {
}

func New() *Repository {
	return &Repository{}
}

func (r Repository) GetPrice(coinTag string) (float64, error) {
	return 0, nil
}

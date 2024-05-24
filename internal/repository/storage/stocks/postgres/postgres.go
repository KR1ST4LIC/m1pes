package postgres

type Repository struct {
}

func New() *Repository {
	return &Repository{}
}

func (r Repository) GetCoinList() ([]string, error) {
	return []string{"биточек", "эфир", "meme"}, nil
}

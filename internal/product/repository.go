package product

type Repository struct {
}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Save() error {
	return nil
}

func (r *Repository) Load() error {
	return nil
}

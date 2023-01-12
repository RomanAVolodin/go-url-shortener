package repositories

type Repository interface {
	GetByID(id string) (string, bool)
	CreateSave(url string) (string, error)
	Save(url, id string) (string, error)
}

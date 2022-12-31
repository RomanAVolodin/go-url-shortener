package repositories

type Repository interface {
	GetByID(id string) (string, bool)
	CreateSave(url string) string
	Save(url, id string)
}

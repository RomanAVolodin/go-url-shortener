package repositories

type Repository interface {
	GetByID(id string) (string, bool)
	Create(url string) string
	CreateBackwardRecord(url, id string)
}

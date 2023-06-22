package redis

type ReferenceStorage struct{}

func NewReferenceStorage() *ReferenceStorage {
	return &ReferenceStorage{}
}

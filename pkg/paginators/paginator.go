package paginators

type Paginator[ItemType any] interface {
	Paginate() error

	GetItems() ([]*ItemType, error)

	HasMore() (bool, error)
}

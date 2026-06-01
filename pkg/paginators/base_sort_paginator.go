package paginators

type BaseSortPaginator[ItemType any] struct {
	*BasePaginator[ItemType, string]
	sortDirection SortDirection
	sortColumn    string
}

func NewBaseSortPaginator[ItemType any](
	filter *PaginationSortDto,
	child Paginator[ItemType],
) *BaseSortPaginator[ItemType] {
	v := &BaseSortPaginator[ItemType]{BasePaginator: NewBasePaginator[ItemType, string](filter, child)}

	if filter == nil {
		return v
	}

	v.sortDirection = *filter.SortDirection
	v.sortColumn = *filter.SortColumn

	return v
}

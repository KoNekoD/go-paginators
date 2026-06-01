package paginators

import (
	"reflect"

	"github.com/KoNekoD/smt/pkg/smt"
	"github.com/pkg/errors"
)

const PaginatorDefaultLimit = 10

type BasePaginator[ItemType any, StartFromT string | int] struct {
	items     []*ItemType
	order     PaginationOrder
	limit     int
	startFrom *StartFromT
	child     Paginator[ItemType]
}

type pagination[StartFromT string | int] interface {
	GetLimit() *int
	GetStartFrom() *StartFromT
	GetOrder() *PaginationOrder
}

func NewBasePaginator[ItemType any, StartFromT string | int](
	filter pagination[StartFromT],
	child Paginator[ItemType],
) *BasePaginator[ItemType, StartFromT] {
	v := &BasePaginator[ItemType, StartFromT]{
		order: PaginationOrderNext,
		limit: PaginatorDefaultLimit,
		child: child,
	}

	if filter == nil || reflect.ValueOf(filter).IsNil() {
		return v
	}

	if filterOrder := filter.GetOrder(); filterOrder != nil {
		v.order = *filterOrder
	}
	if filterLimit := filter.GetLimit(); filterLimit != nil {
		v.limit = *filterLimit
	}
	v.startFrom = filter.GetStartFrom()

	return v
}

func (p *BasePaginator[ItemType, StartFromT]) GetItems() ([]*ItemType, error) {
	items, err := p.getPreloadedItems()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	itemsSlice := smt.SafeSlice(items, 0, p.limit)

	// reverts back sorting during sql request for previous order
	if p.order == PaginationOrderPrev {
		return smt.SliceReverse(itemsSlice), nil
	}

	return itemsSlice, nil
}

func (p *BasePaginator[ItemType, StartFromT]) HasMore() (bool, error) {
	items, err := p.getPreloadedItems()
	if err != nil {
		return false, errors.WithStack(err)
	}

	return len(items) > p.limit, nil
}

func (p *BasePaginator[ItemType, StartFromT]) getPreloadedItems() ([]*ItemType, error) {
	if p.items == nil {
		err := p.child.Paginate()
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	return p.items, nil
}

package paginators

import (
	"fmt"

	"github.com/KoNekoD/gormite/pkg/gormite_query_builders"
	"github.com/pkg/errors"
)

type CursorPaginator[ItemType any] struct {
	*BasePaginator[ItemType, int]
	qb    *gormite_query_builders.QueryBuilder[ItemType]
	alias string
}

func NewCursorPaginator[ItemType any](
	qb *gormite_query_builders.QueryBuilder[ItemType],
	pagination *PaginationDto,
) Paginator[ItemType] {
	v := &CursorPaginator[ItemType]{
		qb:    qb,
		alias: qb.GetRootAliases()[0],
	}

	v.BasePaginator = NewBasePaginator[ItemType, int](pagination, v)

	return v
}

func (p *CursorPaginator[ItemType]) Paginate() error {
	qb := p.qb.Clone()

	orderByKey := fmt.Sprintf("%s.id", p.alias)
	if p.order == PaginationOrderPrev {
		qb.OrderBy(orderByKey, "ASC")
	} else {
		qb.OrderBy(orderByKey, "DESC")
	}

	if p.startFrom != nil {
		if p.order == PaginationOrderPrev {
			qb.AndWhere(fmt.Sprintf("%s.id > :from", p.alias))
		} else {
			qb.AndWhere(fmt.Sprintf("%s.id < :from", p.alias))
		}
		qb.SetParameter("from", *p.startFrom)
	}

	qb.SetMaxResults(p.limit + 1)

	result, err := qb.GetResult()
	if err != nil {
		return errors.WithStack(err)
	}

	p.items = result

	return nil
}

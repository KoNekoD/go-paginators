package paginators

import (
	"context"
	"fmt"

	gdh "github.com/KoNekoD/gormite/pkg/gormite_databases_helpers"
	"github.com/KoNekoD/gormite/pkg/gormite_query_builders"
	"github.com/pkg/errors"
)

type FieldSortNullablePaginator[ItemType any] struct {
	*FieldSortPaginator[ItemType]
}

func NewFieldSortNullablePaginator[ItemType any](
	qb *gormite_query_builders.QueryBuilder[ItemType],
	pagination *PaginationSortDto,
	opts ...FieldSortPaginatorOpt[ItemType],
) *FieldSortNullablePaginator[ItemType] {
	return &FieldSortNullablePaginator[ItemType]{FieldSortPaginator: NewFieldSortPaginator(qb, pagination, opts...)}
}

func (f *FieldSortNullablePaginator[ItemType]) Paginate() error {
	ids, err := f.getIds()
	if err != nil {
		return errors.WithStack(err)
	}

	f.items, err = f.fetchItems(ids)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func (f *FieldSortNullablePaginator[ItemType]) getIds() ([]int, error) {
	qb := f.qb.Clone()

	sortOrder := f.getSortOrderParam()
	sortField := fmt.Sprintf("COALESCE(%s.%s, 0)", f.sortAlias, f.sortColumn)

	if f.startFromId != nil {
		compareOperator := f.getSortCompareOperatorParam()

		where := sortField + " " + compareOperator + " :field " +
			"OR " + sortField + " = :field AND " + f.alias + ".id " + compareOperator + " :id"

		qb.AndWhere(where).SetParameter("id", f.startFromId).SetParameter("field", f.startFromField)
	}

	qb.
		Select(f.alias+".id").
		SetMaxResults(f.limit+1).
		AddOrderBy(sortField, sortOrder).
		AddOrderBy(f.alias+".id", sortOrder)

	sql, err := qb.GetSQL()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return gdh.SelectExecLitSlice[int](context.Background(), qb.Db, sql, qb.GetNamedArgs())
}

func (f *FieldSortNullablePaginator[ItemType]) fetchItems(ids []int) ([]*ItemType, error) {
	qb := f.qb.Clone()

	sortOrder := f.getSortOrderParam()
	sortField := fmt.Sprintf("COALESCE(%s.%s, 0)", f.sortAlias, f.sortColumn)

	qb.
		SetMaxResults(f.limit+1).
		AddOrderBy(sortField, sortOrder).
		AddOrderBy(f.alias+".id", sortOrder)

	if len(ids) > 0 {
		qb.AndWhere(qb.Expr().In(f.alias+".id", qb.PrepareInArgsInt(ids)))
	}

	return qb.GetResult()
}

package paginators

import (
	"context"
	"encoding/json"

	gdh "github.com/KoNekoD/gormite/pkg/gormite_databases_helpers"

	"github.com/KoNekoD/gormite/pkg/gormite_query_builders"
	"github.com/pkg/errors"
)

type FieldSortPaginator[ItemType any] struct {
	*BaseSortPaginator[ItemType]
	qb             *gormite_query_builders.QueryBuilder[ItemType]
	alias          string
	sortAlias      string
	startFromId    *int
	startFromField any
}

type FieldSortPaginatorOpt[ItemType any] func(*FieldSortPaginator[ItemType])

func WithFieldSortPaginatorSortAlias[ItemType any](sortAlias string) FieldSortPaginatorOpt[ItemType] {
	return func(p *FieldSortPaginator[ItemType]) {
		p.sortAlias = sortAlias
	}
}

func NewFieldSortPaginator[ItemType any](
	qb *gormite_query_builders.QueryBuilder[ItemType],
	pagination *PaginationSortDto,
	opts ...FieldSortPaginatorOpt[ItemType],
) *FieldSortPaginator[ItemType] {
	if pagination != nil && pagination.SortDirection == nil {
		defaultOrder := SortDirectionDesc
		pagination.SortDirection = &defaultOrder
	}

	v := &FieldSortPaginator[ItemType]{
		qb:        qb,
		alias:     qb.GetRootAliases()[0],
		sortAlias: qb.GetRootAliases()[0],
	}

	v.BaseSortPaginator = NewBaseSortPaginator[ItemType](pagination, v)

	if v.startFrom != nil {
		compositeStartFrom := make(map[string]any)

		err := json.Unmarshal([]byte(*v.startFrom), &compositeStartFrom)
		if err == nil {
			id, okId := compositeStartFrom["id"]
			field, okField := compositeStartFrom["field"]
			if okId && okField {
				idFloat, ok1 := id.(float64)
				_, ok2 := field.(string)
				_, ok3 := field.(float64)

				if ok1 && (ok2 || ok3) {
					idInt := int(idFloat)
					v.startFromId = &idInt
					v.startFromField = &field
				}
			}
		}
	}

	for _, opt := range opts {
		opt(v)
	}

	return v
}

func (f *FieldSortPaginator[ItemType]) Paginate() error {
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

func (f *FieldSortPaginator[ItemType]) getIds() ([]int, error) {
	qb := f.qb.Clone()

	sortOrder := f.getSortOrderParam()

	if f.startFromId != nil {
		compareOperator := f.getSortCompareOperatorParam()

		where := f.sortAlias + "." + f.sortColumn + " " + compareOperator + " :field " +
			"OR " + f.sortAlias + "." + f.sortColumn + " = :field AND " + f.alias + ".id " + compareOperator + " :id"

		qb.AndWhere(where).SetParameter("id", f.startFromId).SetParameter("field", f.startFromField)
	}

	qb.
		Select(f.alias+".id").
		SetMaxResults(f.limit+1).
		AddOrderBy(f.sortAlias+"."+f.sortColumn, sortOrder).
		AddOrderBy(f.alias+".id", sortOrder)

	sql, err := qb.GetSQL()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return gdh.SelectExecLitSlice[int](context.Background(), qb.Db, sql, qb.GetNamedArgs())
}

func (f *FieldSortPaginator[ItemType]) fetchItems(ids []int) ([]*ItemType, error) {
	qb := f.qb.Clone()

	sortOrder := f.getSortOrderParam()

	qb.
		SetMaxResults(f.limit+1).
		AddOrderBy(f.sortAlias+"."+f.sortColumn, sortOrder).
		AddOrderBy(f.alias+".id", sortOrder)

	if len(ids) > 0 {
		qb.AndWhere(qb.Expr().In(f.alias+".id", qb.PrepareInArgsInt(ids)))
	}

	return qb.GetResult()
}

func (f *FieldSortPaginator[ItemType]) getSortOrderParam() string {
	/** invert order when sort direction is not DESC */
	if f.sortDirection == SortDirectionAsc {
		if f.order == PaginationOrderPrev {
			return "DESC"
		}

		return "ASC"
	}

	if f.order == PaginationOrderPrev {
		return "ASC"
	}

	return "DESC"
}

func (f *FieldSortPaginator[ItemType]) getSortCompareOperatorParam() string {
	/** invert order when sort direction is not DESC */
	if f.sortDirection == SortDirectionAsc {
		if f.order == PaginationOrderPrev {
			return "<"
		}

		return ">"
	}

	if f.order == PaginationOrderPrev {
		return ">"
	}

	return "<"
}

package paginators

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	gdh "github.com/KoNekoD/gormite/pkg/gormite_databases_helpers"

	"github.com/KoNekoD/gormite/pkg/gormite_query_builders"
	"github.com/pkg/errors"
)

type FieldSortPaginator[ItemType any] struct {
	*BaseSortPaginator[ItemType]
	qb             *gormite_query_builders.QueryBuilder[ItemType]
	alias          string
	startFromId    *int
	startFromField any
}

type fieldSortPaginatorOptions struct {
	convertFieldToTime bool
}

type FieldSortPaginatorOption func(*fieldSortPaginatorOptions)

func ConvertFieldToTime() FieldSortPaginatorOption {
	return func(o *fieldSortPaginatorOptions) {
		o.convertFieldToTime = true
	}
}

func NewFieldSortPaginator[ItemType any](
	qb *gormite_query_builders.QueryBuilder[ItemType],
	pagination *PaginationSortDto,
	options ...FieldSortPaginatorOption,
) *FieldSortPaginator[ItemType] {
	opts := &fieldSortPaginatorOptions{}
	for _, fn := range options {
		fn(opts)
	}

	if pagination != nil && pagination.SortDirection == nil {
		defaultOrder := SortDirectionDesc
		pagination.SortDirection = &defaultOrder
	}

	v := &FieldSortPaginator[ItemType]{
		qb:    qb,
		alias: qb.GetRootAliases()[0],
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
				fieldString, ok2 := field.(string)
				_, ok3 := field.(int)

				if ok1 {
					idInt := int(idFloat)
					v.startFromId = &idInt

					if ok2 {
						fieldTime, _ := time.Parse(time.RFC3339, fieldString)
						v.startFromField = fieldTime
					} else if ok3 {
						v.startFromField = &field
					}
				}

			}
		}
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

	rootAlias := qb.GetRootAliases()[0]
	sortOrder := f.getSortOrderParam()

	if f.startFromId != nil {
		compareOperator := f.getSortCompareOperatorParam()

		where := rootAlias + "." + f.sortColumn + " " + compareOperator + " :field " +
			"OR " + rootAlias + "." + f.sortColumn + " = :field AND " + rootAlias + ".id " + compareOperator + " :id"

		qb.AndWhere(where).SetParameter("id", f.startFromId).SetParameter("field", f.startFromField)
	}

	qb.
		Select(rootAlias+".id").
		SetMaxResults(f.limit+1).
		AddOrderBy(fmt.Sprintf("%s.%s, %s.id", rootAlias, f.sortColumn, rootAlias), sortOrder)

	sql, err := qb.GetSQL()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return gdh.SelectExecLitSlice[int](context.Background(), qb.Db, sql, qb.GetNamedArgs())
}

func (f *FieldSortPaginator[ItemType]) fetchItems(ids []int) ([]*ItemType, error) {
	qb := f.qb.Clone()

	rootAlias := qb.GetRootAliases()[0]
	sortOrder := f.getSortOrderParam()

	qb.
		SetMaxResults(f.limit+1).
		AddOrderBy(fmt.Sprintf("%s.%s, %s.id", rootAlias, f.sortColumn, rootAlias), sortOrder)

	if len(ids) > 0 {
		qb.AndWhere(qb.Expr().In(rootAlias+".id", qb.PrepareInArgsInt(ids)))
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

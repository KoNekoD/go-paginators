package paginators

import (
	"github.com/KoNekoD/pgx-colon-query-rewriter/pkg/pgxcqr"

	"github.com/KoNekoD/gormite/pkg/gormite_query_builders"
	"github.com/pkg/errors"
)

type CountCalculator[ResultType any] struct{}

func NewCountCalculator[ResultType any]() *CountCalculator[ResultType] {
	return &CountCalculator[ResultType]{}
}

func (c *CountCalculator[ResultType]) Calculate(
	qb *gormite_query_builders.QueryBuilder[ResultType],
	distinct bool,
) (*CountResponseDto, error) {
	if distinct {
		qb.Distinct(true)
	}

	sql, err := qb.GetSQL()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	sql = "(" + sql + ")"

	args := qb.GetNamedArgs().(pgxcqr.NamedArgs)

	countQb := gormite_query_builders.NewQueryBuilder[int](qb.Db)

	countQb.Select("COUNT(*)").From(sql, "subquery")

	for key, value := range args {
		countQb.SetParameter(key, value)
	}

	var count int

	err = countQb.ExecScanCol(&count)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &CountResponseDto{Count: count}, nil
}

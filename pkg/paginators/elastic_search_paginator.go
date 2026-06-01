package paginators

import (
	"context"
	"strconv"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/pkg/errors"
)

type ElasticSearchPaginator[ItemType any] struct {
	*BasePaginator[ItemType, int]
	index    string
	request  *search.Request
	esClient *elasticsearch.TypedClient
	getByIds func(ids []int, desc bool) ([]*ItemType, error)
}

func NewElasticSearchPaginator[ItemType any](
	pagination *PaginationDto,
	index string,
	req *search.Request,
	esClient *elasticsearch.TypedClient,
	getByIds func(ids []int, desc bool) ([]*ItemType, error),
) *ElasticSearchPaginator[ItemType] {
	v := &ElasticSearchPaginator[ItemType]{
		index:    index,
		request:  req,
		esClient: esClient,
		getByIds: getByIds,
	}

	v.BasePaginator = NewBasePaginator[ItemType, int](pagination, v)

	return v
}

func (e *ElasticSearchPaginator[ItemType]) getSortOrder() *sortorder.SortOrder {
	if e.order == PaginationOrderNext {
		return &sortorder.Desc
	}
	return &sortorder.Asc
}

func (e *ElasticSearchPaginator[ItemType]) applyPagination() {
	e.request.Source_ = types.SourceFilter{Includes: []string{"id"}}

	if e.startFrom != nil {
		e.request.SearchAfter = []types.FieldValue{*e.startFrom}
	}

	size := e.limit + 1
	e.request.Size = &size

	e.request.Sort = []types.SortCombinations{
		types.SortOptions{
			SortOptions: map[string]types.FieldSort{
				"id": {
					Order: e.getSortOrder(),
				},
			},
		},
	}
}

func ElasticSearchExtractIds(resp *search.Response) ([]int, error) {
	var ids []int
	for _, hit := range resp.Hits.Hits {
		if hit.Id_ != nil {
			hitId, err := strconv.Atoi(*hit.Id_)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			ids = append(ids, hitId)
		}
	}
	return ids, nil
}

func (e *ElasticSearchPaginator[ItemType]) Paginate() error {
	ctx := context.Background()

	e.applyPagination()

	resp, err := e.esClient.Search().Index(e.index).Request(e.request).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to execute search")
	}

	ids, err := ElasticSearchExtractIds(resp)
	if err != nil {
		return errors.Wrap(err, "failed to extract ids")
	}

	items, err := e.getByIds(ids, e.order == PaginationOrderNext)
	if err != nil {
		return errors.Wrap(err, "failed to get by ids")
	}

	e.items = items

	return nil
}

package paginators

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/pkg/errors"
)

type ElasticSearchSortPaginator[ItemType any] struct {
	*BaseSortPaginator[ItemType]
	index          string
	request        *search.Request
	esClient       *elasticsearch.TypedClient
	get            func(sources []*ExtractedSource) ([]*ItemType, error)
	startFromId    *int
	startFromField *any
}

func NewElasticSearchSortPaginator[ItemType any](
	pagination *PaginationSortDto,
	index string,
	req *search.Request,
	esClient *elasticsearch.TypedClient,
	get func(sources []*ExtractedSource) ([]*ItemType, error),
) *ElasticSearchSortPaginator[ItemType] {
	if pagination != nil && pagination.SortDirection == nil {
		defaultOrder := SortDirectionDesc
		pagination.SortDirection = &defaultOrder
	}

	v := &ElasticSearchSortPaginator[ItemType]{
		index:    index,
		request:  req,
		esClient: esClient,
		get:      get,
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
				_, ok3 := field.(int)

				if ok1 && (ok2 || ok3) {
					idInt := int(idFloat)
					v.startFromId = &idInt
					v.startFromField = &field
				}
			}
		}
	}

	return v
}

func (e *ElasticSearchSortPaginator[ItemType]) getSortOrder() *sortorder.SortOrder {
	if e.order == PaginationOrderNext {
		return &sortorder.Desc
	}
	return &sortorder.Asc
}

func (e *ElasticSearchSortPaginator[ItemType]) applyPagination() error {
	e.request.Source_ = types.SourceFilter{Includes: []string{"id", e.sortColumn}}

	if e.startFromId != nil {
		e.request.SearchAfter = []types.FieldValue{e.startFromField, e.startFromId}
	}

	size := e.limit + 1
	e.request.Size = &size

	e.request.Sort = []types.SortCombinations{
		types.SortOptions{
			SortOptions: map[string]types.FieldSort{
				e.sortColumn: {
					Order: e.getSortOrder(),
				},
			},
		},
		types.SortOptions{
			SortOptions: map[string]types.FieldSort{
				"id": {
					Order: e.getSortOrder(),
				},
			},
		},
	}

	return nil
}

func (e *ElasticSearchSortPaginator[ItemType]) extractSources(resp *search.Response) ([]*ExtractedSource, error) {
	var sources []*ExtractedSource

	for _, hit := range resp.Hits.Hits {
		var source map[string]any

		err := json.Unmarshal(hit.Source_, &source)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		id := int(source["id"].(float64))

		field := fmt.Sprint(source[e.sortColumn])

		sources = append(sources, &ExtractedSource{Id: id, Field: field})
	}

	return sources, nil
}

func (e *ElasticSearchSortPaginator[ItemType]) Paginate() error {
	ctx := context.Background()

	err := e.applyPagination()
	if err != nil {
		return errors.Wrap(err, "failed to apply pagination")
	}

	resp, err := e.esClient.Search().Index(e.index).Request(e.request).Do(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to execute search")
	}

	sources, err := e.extractSources(resp)
	if err != nil {
		return errors.Wrap(err, "failed to extractSources sources")
	}

	items, err := e.get(sources)
	if err != nil {
		return errors.Wrap(err, "failed to get by sources")
	}

	e.items = items

	return nil
}

package paginators

type PaginationDto struct {
	Limit     *int             `json:"limit" binding:"omitnil,oneof=10 25 50"`
	StartFrom *int             `json:"startFrom" binding:"omitnil,gt=0"`
	Order     *PaginationOrder `json:"order" binding:"omitnil,oneof=next prev"`
}

func (p *PaginationDto) GetLimit() *int {
	return p.Limit
}

func (p *PaginationDto) GetStartFrom() *int {
	return p.StartFrom
}

func (p *PaginationDto) GetOrder() *PaginationOrder {
	return p.Order
}

type PaginationSortDto struct {
	Limit         *int             `json:"limit" binding:"omitnil,oneof=10 25 50"`
	StartFrom     *string          `json:"startFrom" binding:"omitnil"`
	Order         *PaginationOrder `json:"order" binding:"omitnil,oneof=next prev"`
	SortDirection *SortDirection   `json:"sort" binding:"omitnil,oneof=asc desc"`
	SortColumn    *string          `json:"sortColumn" binding:"omitnil"`
}

func (p *PaginationSortDto) GetLimit() *int {
	return p.Limit
}

func (p *PaginationSortDto) GetStartFrom() *string {
	return p.StartFrom
}

func (p *PaginationSortDto) GetOrder() *PaginationOrder {
	return p.Order
}

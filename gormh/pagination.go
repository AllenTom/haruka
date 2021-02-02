package gormh

type PageFilter interface {
	SetPageFilter(page int, pageSize int)
	GetOffset() int
	GetLimit() int
}
type DefaultPageFilter struct {
	Page     int
	PageSize int
}

func (b *DefaultPageFilter) SetPageFilter(page int, pageSize int) {
	b.Page = page
	b.PageSize = pageSize
}

func (b *DefaultPageFilter) GetOffset() int {
	return b.PageSize * (b.Page - 1)
}

func (b *DefaultPageFilter) GetLimit() int {
	if b.PageSize == 0 {
		return 1
	} else {
		return b.PageSize
	}
}

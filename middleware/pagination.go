package middleware

import (
	"github.com/allentom/haruka"
)

type PaginationMiddleware struct {
	pageSizeLookUp  string
	pageLookUp      string
	defaultPage     int
	defaultPageSize int
}

func (p *PaginationMiddleware) OnRequest(ctx *haruka.Context) {
	page, err := ctx.GetQueryInt(p.pageLookUp)
	if err != nil {
		page = p.defaultPage
	}
	pageSize, err := ctx.GetQueryInt(p.pageSizeLookUp)
	if err != nil {
		pageSize = p.defaultPageSize
	}
	ctx.Param["page"] = page
	ctx.Param["pageSize"] = pageSize
}

func NewPaginationMiddleware(pageLookUp string, pageSizeLookUp string, defaultPage int, defaultPageSize int) *PaginationMiddleware {
	return &PaginationMiddleware{
		pageSizeLookUp:  pageSizeLookUp,
		pageLookUp:      pageLookUp,
		defaultPage:     defaultPage,
		defaultPageSize: defaultPageSize,
	}
}

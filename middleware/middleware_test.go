package middleware

import (
	"github.com/allentom/haruka"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
)

func TestPaginationMiddleware_OnRequest(t *testing.T) {
	url, _ := url.Parse("http://localhost:8090/ping?page=1&pageSize=20")
	req := &http.Request{URL: url}
	ctx := &haruka.Context{
		Request: req,
		Param:   map[string]interface{}{},
	}
	middleware := PaginationMiddleware{
		pageSizeLookUp: "pageSize",
		pageLookUp:     "page",
	}
	middleware.OnRequest(ctx)
	assert.Equal(t, ctx.Param["page"].(int), 1)
	assert.Equal(t, ctx.Param["pageSize"].(int), 20)

}

package serializer

import (
	"math"
	"net/url"
	"strconv"
)

func getNextPageURL(url *url.URL, count int64, page int, pageSize int) string {
	totalPage := math.Ceil(float64(count) / float64(pageSize))
	query := url.Query()
	if totalPage > float64(page) {
		query.Set("page", strconv.Itoa(page+1))
		url.RawQuery = query.Encode()
		return url.String()
	} else {
		return ""
	}
}

func getNextPreviousURL(url *url.URL, count int64, page int) string {
	query := url.Query()
	if page > 2 {
		query.Set("page", strconv.Itoa(page-1))
		url.RawQuery = query.Encode()
		return url.String()
	} else {
		return ""
	}
}

type ListContainerSerializer interface {
	SerializeList(result interface{}, context map[string]interface{})
}

type DefaultListContainer struct {
	Count    int64       `json:"count"`
	Next     string      `json:"next"`
	Previous string      `json:"previous"`
	Page     int         `json:"page"`
	PageSize int         `json:"pageSize"`
	Results  interface{} `json:"result"`
}

func (c *DefaultListContainer) SerializeList(result interface{}, context map[string]interface{}) {
	page := context["page"].(int)
	pageSize := context["pageSize"].(int)
	requestUrl := context["url"].(*url.URL)
	count := context["count"].(int64)
	c.Count = count
	c.Next = getNextPageURL(requestUrl, count, page, pageSize)
	c.Previous = getNextPreviousURL(requestUrl, count, page)
	c.Results = result
	c.PageSize = pageSize
	c.Page = page
	return
}

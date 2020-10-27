package Haruka

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Parser interface {
	Parse(r *http.Request, target interface{}) error
}
type JSONParser struct {
}

func (p *JSONParser) Parse(r *http.Request, target interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, target)
}

func (c *Context) ParseJson(target interface{}) error {
	parser := JSONParser{}
	return parser.Parse(c.Request, target)
}

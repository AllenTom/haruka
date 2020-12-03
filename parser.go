package haruka

import (
	"encoding/json"
	"encoding/xml"
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

type XMLParser struct {
}

func (p *XMLParser) Parse(r *http.Request, target interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(body, target)
}

func (c *Context) ParseXML(target interface{}) error {
	parser := XMLParser{}
	return parser.Parse(c.Request, target)
}

package haruka

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type TestJsonBody struct {
	Say string `json:"say"`
}

func TestJSONParser_Parse(t *testing.T) {
	testRequest, _ := http.NewRequest("POST", "/", bytes.NewBufferString("{\"say\":\"hello\"}"))
	context := Context{
		Writer:  nil,
		Request: testRequest,
	}
	body := &TestJsonBody{}
	err := context.ParseJson(body)
	if err != nil {
		assert.Error(t, err)
	}
	assert.Equal(t, "hello", body.Say)
}

type TestXMLBody struct {
	Say string `xml:"say"`
}

func TestXMLParser_Parse(t *testing.T) {
	testRequest, _ := http.NewRequest("POST", "/", bytes.NewBufferString("<test><say>hello</say></test>"))
	context := Context{
		Writer:  nil,
		Request: testRequest,
	}
	body := &TestXMLBody{}
	err := context.ParseXML(body)
	if err != nil {
		assert.Error(t, err)
	}
	assert.Equal(t, "hello", body.Say)
}

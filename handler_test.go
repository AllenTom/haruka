package haruka

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestReadPathParameter(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping/{key}/{key2}/key4/{key3}", func(context *Context) {
		err := context.XML(XMLBody{Say: "Hello"})
		assert.Equal(t, "123asd", context.Parameters["key"])
		assert.Equal(t, "123", context.Parameters["key2"])
		assert.Equal(t, "asd", context.Parameters["key3"])
		if err != nil {
			t.Error(e)
		}
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	testRequest(t, "http://localhost:8090/ping/123asd/123/key4/asd?key=233", "<test><say>Hello</say></test>")
}

func TestContext_GetQueryInt(t *testing.T) {
	url, _ := url.Parse("http://localhost:8090/ping/123asd/123/key4/asd?key=233")
	req := &http.Request{URL: url}
	ctx := Context{Request: req}
	key, err := ctx.GetQueryInt("key")
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, 233, key)
}

func TestContext_GetQueryString(t *testing.T) {
	url, _ := url.Parse("http://localhost:8090/ping/123asd/123/key4/asd?key=233")
	req := &http.Request{URL: url}
	ctx := Context{Request: req}
	key := ctx.GetQueryString("key")
	assert.Equal(t, "233", key)
}

func TestContext_GetPathParameterAsString(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping/{key}", func(context *Context) {
		err := context.XML(XMLBody{Say: "Hello"})
		assert.Equal(t, "123asd", context.GetPathParameterAsString("key"))
		if err != nil {
			t.Error(e)
		}
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	<-time.After(3 * time.Second)
	testRequest(t, "http://localhost:8090/ping/123asd", "<test><say>Hello</say></test>")
}

func TestContext_GetPathParameterAsInt(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping/{key}", func(context *Context) {
		err := context.XML(XMLBody{Say: "Hello"})
		value, err := context.GetPathParameterAsInt("key")
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, 123, value)
		if err != nil {
			t.Error(e)
		}
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	testRequest(t, "http://localhost:8090/ping/123", "<test><say>Hello</say></test>")
}

func TestContext_GetQueryStrings(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping", func(context *Context) {
		err := context.XML(XMLBody{Say: "Hello"})
		key := context.GetQueryStrings("key")
		key2 := context.GetQueryStrings("key2")
		key3 := context.GetQueryStrings("key3")
		assert.Equal(t, []string{"123", "1234"}, key)
		assert.Equal(t, []string{"123"}, key2)
		assert.Nil(t, key3)
		if err != nil {
			t.Error(e)
		}
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	testRequest(t, "http://localhost:8090/ping?key=123&key=1234&key2=123", "<test><say>Hello</say></test>")
}

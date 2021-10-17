package haruka

import (
	"fmt"
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

type Input struct {
	UintId                    uint          `hsource:"query" hname:"id"`
	Int64Id                   int64         `hsource:"query" hname:"id"`
	Name                      string        `hsource:"query" hname:"name"`
	Token                     string        `hsource:"path" hname:"token"`
	NumberItems               []int64       `hsource:"query" hname:"nitems"`
	NumberItemsArray          [5]int64      `hsource:"query" hname:"nitems"`
	NumberStringItemsArray    []string      `hsource:"query" hname:"nitems"`
	NumberInterfaceItemsArray []interface{} `hsource:"query" hname:"nitems"`
	Again                     struct {
		UintId                    uint          `hsource:"query" hname:"id"`
		Int64Id                   int64         `hsource:"query" hname:"id"`
		Name                      string        `hsource:"query" hname:"name"`
		Token                     string        `hsource:"path" hname:"token"`
		NumberItems               []int64       `hsource:"query" hname:"nitems"`
		NumberItemsArray          [5]int64      `hsource:"query" hname:"nitems"`
		NumberStringItemsArray    []string      `hsource:"query" hname:"nitems"`
		NumberInterfaceItemsArray []interface{} `hsource:"query" hname:"nitems"`
		AAgain                    struct {
			UintId                    uint          `hsource:"query" hname:"id"`
			Int64Id                   int64         `hsource:"query" hname:"id"`
			Name                      string        `hsource:"query" hname:"name"`
			Token                     string        `hsource:"path" hname:"token"`
			NumberItems               []int64       `hsource:"query" hname:"nitems"`
			NumberItemsArray          [5]int64      `hsource:"query" hname:"nitems"`
			NumberStringItemsArray    []string      `hsource:"query" hname:"nitems"`
			NumberInterfaceItemsArray []interface{} `hsource:"query" hname:"nitems"`
		}
	}
	Param                 string     `hsource:"param" hname:"name"`
	ParamInt              int64      `hsource:"param" hname:"id"`
	ParamSlice            []string   `hsource:"param" hname:"slice"`
	StartTime             *time.Time `hsource:"query" hname:"startTime" format:"2006-01-02"`
	FormValue1            int        `hsource:"form" hname:"value1"`
	FormValue2            string     `hsource:"form" hname:"value2"`
	FormStringSliceValue2 []string   `hsource:"form" hname:"value3"`
	FormIntSliceValue2    []int      `hsource:"form" hname:"value4"`
}

func TestConvert(t *testing.T) {
	url, _ := url.Parse("http://localhost:8090/ping?id=1&id=20&name=aren&nitems=1&&nitems=2&&nitems=3&startTime=2021-01-01")
	req := &http.Request{URL: url}
	req.Form = map[string][]string{
		"value1": {"1"},
		"value2": {"textValue"},
		"value3": {"elm1", "elm2"},
		"value4": {"999", "1000"},
	}

	ctx := &Context{
		Request: req,
		Param: map[string]interface{}{
			"name":  "name",
			"id":    int64(1),
			"slice": []string{"1", "2", "3"},
		},
		Parameters: map[string]string{
			"token": "path_token",
		},
	}
	input := &Input{}
	err := ctx.BindingInput(input)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, int64(1), input.Int64Id)
	assert.Equal(t, uint(1), input.UintId)
	assert.Equal(t, "aren", input.Name)
	assert.Equal(t, "path_token", input.Token)
	for idx, num := range input.NumberItems {
		assert.Equal(t, int64(idx+1), num)
	}

	assert.Equal(t, int64(1), input.Again.Int64Id)
	assert.Equal(t, uint(1), input.Again.UintId)
	assert.Equal(t, "aren", input.Again.Name)
	assert.Equal(t, "path_token", input.Again.Token)
	for idx, num := range input.Again.NumberItems {
		assert.Equal(t, int64(idx+1), num)
	}

	assert.Equal(t, "name", input.Param)
	assert.Equal(t, int64(1), input.ParamInt)
	for idx, num := range input.ParamSlice {
		assert.Equal(t, fmt.Sprintf("%d", idx+1), num)
	}
}

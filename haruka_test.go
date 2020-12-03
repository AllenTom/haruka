package haruka

import (
	"crypto/tls"
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
)

func testRequest(t *testing.T, url string, expected string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, ioerr := ioutil.ReadAll(resp.Body)
	assert.NoError(t, ioerr)
	assert.Equal(t, expected, string(body), "resp body should match")
	assert.Equal(t, "200 OK", resp.Status, "should get a 200")
}
func TestRunAndListen(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping", func(context *Context) {
		data := map[string]interface{}{
			"say": "pong",
		}
		err := context.JSON(data)
		if err != nil {
			t.Error(e)
		}
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	testRequest(t, "http://localhost:8090/ping", "{\"say\":\"pong\"}")
}

type XMLBody struct {
	XMLName xml.Name `xml:"test"`
	Say     string   `xml:"say"`
}

func TestXML(t *testing.T) {
	e := NewEngine()
	e.Router.AddHandler("/ping", func(context *Context) {
		err := context.XML(XMLBody{Say: "Hello"})
		if err != nil {
			t.Error(e)
		}
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	testRequest(t, "http://localhost:8090/ping", "<test><say>Hello</say></test>")
}

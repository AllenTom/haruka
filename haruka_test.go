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
func testMakePOSTRequest(t *testing.T, url string) ([]byte, *http.Response) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Post(url, "application/json", nil)
	assert.NoError(t, err)
	defer resp.Body.Close()

	body, ioerr := ioutil.ReadAll(resp.Body)
	assert.NoError(t, ioerr)
	return body, resp
}
func testMakeGETRequest(t *testing.T, url string) ([]byte, *http.Response) {
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
	return body, resp
}
func testMakePUTRequest(t *testing.T, url string) ([]byte, *http.Response) {
	client := &http.Client{}
	request, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		t.Error(err)
	}
	response, err := client.Do(request)
	body, ioerr := ioutil.ReadAll(response.Body)
	assert.NoError(t, ioerr)
	return body, response
}

func testMakePATCHRequest(t *testing.T, url string) ([]byte, *http.Response) {
	client := &http.Client{}
	request, err := http.NewRequest("PATCH", url, nil)
	if err != nil {
		t.Error(err)
	}
	response, err := client.Do(request)
	body, ioerr := ioutil.ReadAll(response.Body)
	assert.NoError(t, ioerr)
	return body, response
}

func testMakeDELETERequest(t *testing.T, url string) ([]byte, *http.Response) {
	client := &http.Client{}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		t.Error(err)
	}
	response, err := client.Do(request)
	body, ioerr := ioutil.ReadAll(response.Body)
	assert.NoError(t, ioerr)
	return body, response
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

func TestRouter_GET(t *testing.T) {
	e := NewEngine()
	e.Router.GET("/ping", func(context *Context) {
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
	_, resp := testMakePOSTRequest(t, "http://localhost:8090/ping")
	testRequest(t, "http://localhost:8090/ping", "{\"say\":\"pong\"}")
	// other methods 405 response
	assert.Equal(t, resp.StatusCode, http.StatusMethodNotAllowed)
}

func TestRouter_POST(t *testing.T) {
	e := NewEngine()
	e.Router.POST("/ping", func(context *Context) {
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

	body, _ := testMakePOSTRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", string(body))
	// other methods 405 response
	_, postResponse := testMakeGETRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, postResponse.StatusCode, http.StatusMethodNotAllowed)
}

func TestRouter_PUT(t *testing.T) {
	e := NewEngine()
	e.Router.PUT("/ping", func(context *Context) {
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

	body, _ := testMakePUTRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", string(body))
	// other methods 405 response
	_, postResponse := testMakePOSTRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, postResponse.StatusCode, http.StatusMethodNotAllowed)
}

func TestRouter_PATCH(t *testing.T) {
	e := NewEngine()
	e.Router.PATCH("/ping", func(context *Context) {
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

	body, _ := testMakePATCHRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", string(body))
	// other methods 405 response
	_, postResponse := testMakePOSTRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, postResponse.StatusCode, http.StatusMethodNotAllowed)
}

func TestRouter_DELETE(t *testing.T) {
	e := NewEngine()
	e.Router.DELETE("/ping", func(context *Context) {
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

	body, _ := testMakeDELETERequest(t, "http://localhost:8090/ping")
	assert.Equal(t, "{\"say\":\"pong\"}", string(body))
	// other methods 405 response
	_, postResponse := testMakePOSTRequest(t, "http://localhost:8090/ping")
	assert.Equal(t, postResponse.StatusCode, http.StatusMethodNotAllowed)
}

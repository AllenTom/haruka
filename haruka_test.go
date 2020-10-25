package Haruka

import (
	"crypto/tls"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
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
		rawData, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		context.Writer.Header().Set("Content-Type", "application/json")
		context.Writer.Write(rawData)
	})
	go func() {
		e.RunAndListen(":8090")
	}()
	<-time.After(3 * time.Second)
	testRequest(t, "http://localhost:8090/ping", "{\"say\":\"pong\"}")
}

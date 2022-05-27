# haruka
web framework

## Example

```go
package main

import "github.com/allentom/haruka"

func main() {
	e := haruka.NewEngine()
	e.Router.GET("/ping", func(context *Context) {
		context.JSONWithStatus(haruka.JSON{
			"message":"Pong",
		}, http.StatusOK)
    })
	e.ListenAndServe(":8080")
}
```
## Router
### With specific path

```go
e.Router.GET("/ping", <YourHandler>)
// with path parameters
e.Router.GET("/ping/{name}", <YourHandler>)
// with regex
e.Router.GET("/ping/{name:[0-9|a-z|A-Z]+}", <YourHandler>)
```
### With different methods

```go
e.Router.GET("/ping", <YourHandler>)
e.Router.POST("/ping", <YourHandler>)
e.Router.PATCH("/ping", <YourHandler>)
e.Router.Put("/ping", <YourHandler>)
e.Router.OPTION("/ping", <YourHandler>)

// for all methods
e.Router.AddHandler("/ping", <YourHandler>)
// for multiple methods
e.Router.METHODS("/ping", []string{"GET","POST"}, <YourHandler>)
```

## Handler

### Get query parameters

```go
func (c *Context) {
    context.GetQueryString("id")
    context.GetQueryStrings("ids")
    context.GetQueryInt("id")
}
```

### Get path parameters

```go
func (c *Context) {
    context.GetPathParameterAsInt("id")
    context.GetPathParameterAsString("id")
}

```

package templates

var RestTemplate = `package application

import "github.com/allentom/haruka"

type Create{{ .modelName }}RequestBody struct {

}
var Create{{ .modelName }}Handler haruka.RequestHandler  = func(context *haruka.Context) {
	var requestBody Create{{ .modelName }}RequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		// error handling
	}
	// custom code there
}

var Delete{{ .modelName }}Handler haruka.RequestHandler  = func(context *haruka.Context) {
	id,err := context.GetPathParameterAsInt("id")
	if err != nil {

	}
	// custom code there
}

var Get{{ .modelName }}ListHandler haruka.RequestHandler = func(context *haruka.Context) {

}

type Update{{ .modelName }}RequestBody struct {

}
var Update{{ .modelName }}Handler haruka.RequestHandler = func(context *haruka.Context) {
	var requestBody Update{{ .modelName }}RequestBody
	err := context.ParseJson(&requestBody)
	if err != nil {
		// error handling
	}
}
`

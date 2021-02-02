package blueprint

import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/gormh"
	"github.com/allentom/haruka/serializer"
	"github.com/allentom/haruka/validator"
	"github.com/jinzhu/copier"
	"net/http"
	"reflect"
	"strconv"
)

// CreateModelView create model view
type CreateModelView struct {
	Context          *haruka.Context
	CreateModel      func() interface{}
	ResponseTemplate serializer.TemplateSerializer
	RequestBody      interface{}
	OnBeforeCreate   func(v *CreateModelView, modelToCreate interface{})
	OnError          func(err error)
	GetValidators    func(v *CreateModelView) []validator.Validator
	OnCreate         func(view *CreateModelView, model interface{}) (interface{}, error)
}

func (v *CreateModelView) Run() {
	var err error
	err = v.Context.ParseJson(v.RequestBody)
	if err != nil {
		return
	}
	if v.GetValidators != nil {
		err = validator.RunValidators(v.GetValidators(v)...)
		if err != nil {
			v.OnError(err)
			return
		}
	}
	createModel := v.CreateModel()
	err = copier.Copy(createModel, v.RequestBody)
	if err != nil {
		v.OnError(err)
		return
	}
	if v.OnBeforeCreate != nil {
		v.OnBeforeCreate(v, createModel)
	}
	model, err := v.OnCreate(v, createModel)
	if err != nil {
		v.OnError(err)
		return
	}
	//serializer response

	err = v.ResponseTemplate.Serializer(model, map[string]interface{}{})
	if err != nil {
		v.OnError(err)
		return
	}
	v.Context.Writer.Header().Set("Content-Type", "application/json")
	v.Context.Writer.WriteHeader(http.StatusCreated)
	v.Context.JSON(v.ResponseTemplate)
}

type FilterMapping struct {
	Lookup string
	Method string
	Many   bool
}

type PageReader interface {
	Read(ctx *haruka.Context) (int, int)
}

type DefaultPagination struct {
}

func (p *DefaultPagination) Read(ctx *haruka.Context) (int, int) {
	return ctx.Param["page"].(int), ctx.Param["pageSize"].(int)
}

// ListView fetch models list view
type ListView struct {
	Context              *haruka.Context
	Pagination           PageReader
	QueryBuilder         interface{}
	FilterMapping        []FilterMapping
	GetSerializerContext func(v *ListView, result interface{}) map[string]interface{}
	GetTemplate          func() serializer.TemplateSerializer
	GetContainer         func() serializer.ListContainerSerializer
	OnApplyQuery         func()
	OnError              func(err error)
}

func (v *ListView) Run() {
	if v.Pagination == nil {
		v.Pagination = &DefaultPagination{}
	}
	page, PageSize := v.Pagination.Read(v.Context)

	pageFilter := (v.QueryBuilder).(gormh.PageFilter)
	pageFilter.SetPageFilter(page, PageSize)

	for _, filter := range v.FilterMapping {
		FilterByParam(v.Context, filter.Lookup, v.QueryBuilder, filter.Method, filter.Many)
	}
	if v.OnApplyQuery != nil {
		v.OnApplyQuery()
	}
	modelsReader := (v.QueryBuilder).(gormh.ModelsReader)
	count, modelList, err := modelsReader.ReadModels()
	if err != nil {
		v.OnError(err)
		return
	}
	serializerContext := map[string]interface{}{}
	if v.GetSerializerContext != nil {
		serializerContext = v.GetSerializerContext(v, modelList)
	}

	result := serializer.SerializeMultipleTemplate(modelList, v.GetTemplate(), serializerContext)
	responseBody := v.GetContainer()
	responseBody.SerializeList(result, map[string]interface{}{
		"page":     page,
		"pageSize": PageSize,
		"count":    count,
		"url":      v.Context.Request.URL,
	})
	v.Context.JSON(responseBody)
}

type DeleteModelView struct {
	Context         *haruka.Context
	Lookup          string
	OnError         func(err error)
	Model           gormh.DataModel
	GetResponseBody func() interface{}
}

func (v *DeleteModelView) Run() {
	rawId := v.Context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		v.OnError(err)
		return
	}
	err = v.Model.DeleteById(uint(id))
	if err != nil {
		v.OnError(err)
	}
	v.Context.Writer.Header().Set("Content-Type", "application/json")
	v.Context.Writer.WriteHeader(http.StatusOK)
	if v.GetResponseBody != nil {
		responseBody := v.GetResponseBody
		v.Context.JSON(responseBody)
		return
	}
	v.Context.JSON(haruka.JSON{
		"success": true,
	})

}

type UpdateModelView struct {
	Context  *haruka.Context
	Lookup   string
	OnError  func(err error)
	Model    gormh.DataModel
	Template serializer.TemplateSerializer
}

func (v *UpdateModelView) Run() {
	rawId := v.Context.Parameters["id"]
	id, err := strconv.Atoi(rawId)
	if err != nil {
		v.OnError(err)
		return
	}
	requestBody := make(map[string]interface{}, 0)
	err = v.Context.ParseJson(&requestBody)
	if err != nil {
		v.OnError(err)
		return
	}
	model, err := v.Model.UpdateById(uint(id), requestBody)
	if err != nil {
		v.OnError(err)
		return
	}
	err = v.Template.Serializer(model, map[string]interface{}{})
	if err != nil {
		v.OnError(err)
		return
	}
	v.Context.Writer.Header().Set("Content-Type", "application/json")
	v.Context.Writer.WriteHeader(http.StatusOK)
	v.Context.JSON(v.Template)

}

// FilterByParam binding query parameters to query builder
func FilterByParam(controller *haruka.Context, name string, queryBuilder interface{}, methodName string, many bool) {
	params := controller.GetQueryStrings(name)
	if params == nil || len(params) == 0 {
		return
	}
	builderRef := reflect.ValueOf(queryBuilder)
	filterMethodRef := builderRef.MethodByName(methodName)
	inputs := make([]reflect.Value, len(params))
	if !many {
		inputs[0] = reflect.ValueOf(params[0])
	} else {
		for i := range params {
			inputs[i] = reflect.ValueOf(params[i])
		}
	}
	filterMethodRef.Call(inputs)
}

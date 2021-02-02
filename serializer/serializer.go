package serializer

import (
	"github.com/jinzhu/copier"
	"reflect"
)

type TemplateSerializer interface {
	Serializer(dataModel interface{}, context map[string]interface{}) error
}

type DefaultTemplateSerializer struct {
}

func (s *DefaultTemplateSerializer) Serializer(model interface{}) error {
	return copier.Copy(s, model)
}

func SerializeMultipleTemplate(items interface{}, template TemplateSerializer, context map[string]interface{}) interface{} {
	result := make([]interface{}, 0)
	itemListRef := reflect.ValueOf(items)
	for itemIdx := 0; itemIdx < itemListRef.Len(); itemIdx++ {
		itemTemplate := reflect.New(reflect.TypeOf(template).Elem())
		tmp := itemTemplate.Interface().(TemplateSerializer)
		tmp.Serializer(itemListRef.Index(itemIdx).Interface(), context)
		result = append(result, tmp)
	}
	return result
}

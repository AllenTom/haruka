package haruka

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
	"time"
)

var (
	timeTypeKind      reflect.Kind
	defaultTimeFormat = "2006-01-02 15:04:05"
)

func init() {
	timeType := time.Now()
	timeTypeKind = reflect.TypeOf(&timeType).Kind()
}

type Context struct {
	Writer      http.ResponseWriter
	Request     *http.Request
	Parameters  map[string]string
	Param       map[string]interface{}
	isAbort     bool
	isInterrupt bool
}
type RequestHandler func(context *Context)

// get query string from url
func (c *Context) GetQueryString(key string) string {
	return c.Request.URL.Query().Get(key)
}

// get query string as []string
func (c *Context) GetQueryStrings(key string) []string {
	return c.Request.URL.Query()[key]
}

// get query string as int
func (c *Context) GetQueryInt(key string) (int, error) {
	rawValue := c.Request.URL.Query().Get(key)
	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// get path parameter as string
func (c *Context) GetPathParameterAsString(key string) string {
	return c.Parameters[key]
}

// get path parameter as int
func (c *Context) GetPathParameterAsInt(key string) (int, error) {
	rawValue := c.Parameters[key]
	value, err := strconv.Atoi(rawValue)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// interrupt middleware chain
func (c *Context) Abort() {
	c.isAbort = true
}

// interrupt middleware chain
func (c *Context) Interrupt() {
	c.isInterrupt = true
}

type FormFile struct {
	Header *multipart.FileHeader
	File   multipart.File
}

func setValue(value reflect.Value, rawValue string) error {
	switch value.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := strconv.ParseInt(rawValue, 10, 64)
		if err != nil {
			return err
		}
		value.SetInt(v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := strconv.ParseUint(rawValue, 10, 64)
		if err != nil {
			return err
		}
		value.SetUint(v)
	case reflect.String:
		value.SetString(rawValue)
	case reflect.Interface:
		value.Set(reflect.ValueOf(rawValue))
	default:
		return errors.New("unknown type")
	}
	return nil
}
func BindingValue(valueField reflect.Value, rawValue []string, tags reflect.StructTag) error {
	var err error
	if valueField.Kind() == reflect.Slice || valueField.Kind() == reflect.Array {
		if valueField.Kind() == reflect.Array && valueField.Cap() < len(rawValue) {
			return errors.New("array cap insufficient")
		}
		if valueField.Kind() == reflect.Slice {
			valueField.Set(reflect.MakeSlice(valueField.Type(), len(rawValue), len(rawValue)))
		}
		for idx, s := range rawValue {
			element := valueField.Index(idx)
			err = setValue(element, s)
			if err != nil {
				return err
			}
		}
	} else if valueField.Kind() == timeTypeKind {
		timeFormat := tags.Get("format")
		if len(timeFormat) == 0 {
			timeFormat = defaultTimeFormat
		}
		if len(rawValue[0]) == 0 {
			return nil
		}
		timeValue, err := time.Parse(timeFormat, rawValue[0])
		if err != nil {
			return err
		}
		valueField.Set(reflect.ValueOf(&timeValue))
		fmt.Println("time")
	} else {
		// not iteration
		err = setValue(valueField, rawValue[0])
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	return nil
}
func bindingWalk(c *Context, v reflect.Value) error {
	var err error
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		valueField := v.Field(i)

		if valueField.Kind() == reflect.Struct {
			err = bindingWalk(c, valueField)
			if err != nil {
				return err
			}
		}

		tags := field.Tag
		source := tags.Get("hsource")
		sourceName := tags.Get("hname")
		switch source {
		case "query":
			err = BindingValue(valueField, c.GetQueryStrings(sourceName), tags)
			if err != nil {
				return err
			}
		case "path":
			if valueField.Kind() == reflect.Slice || valueField.Kind() == reflect.Array {

			} else {
				rawValue := c.GetPathParameterAsString(sourceName)
				err = setValue(valueField, rawValue)
				if err != nil {
					return err
				}
			}
		case "param":
			rawData := c.Param[sourceName]
			valueField.Set(reflect.ValueOf(rawData))
		case "form":
			if formFileValue, ok := valueField.Interface().(FormFile); ok {
				formFileValue = FormFile{}
				formFileValue.File, formFileValue.Header, err = c.Request.FormFile(sourceName)
				if err != nil {
					return err
				}

			} else if formFilesValue, ok := valueField.Interface().([]FormFile); ok {
				formFilesValue = []FormFile{}
				headers := c.Request.MultipartForm.File[sourceName]
				if headers == nil {
					continue
				}
				for _, header := range headers {
					formFile, err := header.Open()
					if err != nil {
						return err
					}
					formFilesValue = append(formFilesValue, FormFile{
						Header: header,
						File:   formFile,
					})
				}
			} else {
				rawValue := c.Request.Form[sourceName]
				if len(rawValue) == 0 {
					continue
				}
				err = BindingValue(valueField, rawValue, tags)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
func (c *Context) BindingInput(input interface{}) error {
	var err error
	v := reflect.ValueOf(input).Elem()
	err = bindingWalk(c, v)
	if err != nil {
		return err
	}
	return nil
}

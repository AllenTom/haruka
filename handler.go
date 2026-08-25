package haruka

import (
	"errors"
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

// Context for request
type Context struct {
	// Pattern matched url path
	Pattern string
	// The http.ResponseWriter object.
	Writer http.ResponseWriter
	// The http.Request object.
	Request *http.Request
	// Path params
	Parameters map[string]string
	// For middleware to store dara
	Param map[string]interface{}
	// Abort flag the request.
	isAbort   bool
	recorder  *responseRecorder
	startTime time.Time
}
type RequestHandler func(context *Context)

// StartTime returns the moment the request entered the framework.
func (c *Context) StartTime() time.Time {
	return c.startTime
}

// StatusCode returns the HTTP status code written to the response,
// or 0 when nothing has been written yet.
func (c *Context) StatusCode() int {
	if c.recorder == nil {
		return 0
	}
	return c.recorder.Status()
}

// RawWriter exposes the original http.ResponseWriter before any
// framework wrapping. Needed when integration code must satisfy
// unexported interfaces of net/http (e.g. MaxBytesReader's 413 hook).
func (c *Context) RawWriter() http.ResponseWriter {
	if c.recorder == nil {
		return c.Writer
	}
	return c.recorder.ResponseWriter
}

// ClientIP resolves the client address honouring common proxy headers.
func (c *Context) ClientIP() string {
	if v := c.Request.Header.Get("X-Forwarded-For"); v != "" {
		if idx := indexByteString(v, ','); idx >= 0 {
			v = v[:idx]
		}
		return trimSpaceString(v)
	}
	if v := c.Request.Header.Get("X-Real-IP"); v != "" {
		return v
	}
	host := c.Request.RemoteAddr
	if idx := indexByteString(host, ':'); idx >= 0 {
		host = host[:idx]
	}
	return host
}

func indexByteString(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func trimSpaceString(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

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

// AbortWithStatus writes the status code and aborts the chain.
func (c *Context) AbortWithStatus(status int) {
	c.recorder.WriteHeader(status)
	c.Abort()
}

// AbortWithStatusJSON writes data as JSON with the given status and
// aborts the chain.
func (c *Context) AbortWithStatusJSON(status int, data interface{}) {
	_ = c.JSONWithStatus(data, status)
	c.Abort()
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
	case reflect.Float32, reflect.Float64:
		v, err := strconv.ParseFloat(rawValue, 64)
		if err != nil {
			return err
		}
		value.SetFloat(v)
	case reflect.Bool:
		v, err := strconv.ParseBool(rawValue)
		if err != nil {
			return err
		}
		value.SetBool(v)
	case timeTypeKind:
		timeFormat := defaultTimeFormat
		v, err := time.Parse(timeFormat, rawValue)
		if err != nil {
			return err
		}
		value.Set(reflect.ValueOf(v))

	default:
		return errors.New("unknown type")
	}
	return nil
}
func BindingValue(valueField reflect.Value, rawValue []string, tags reflect.StructTag) error {
	if rawValue == nil {
		return nil
	}
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

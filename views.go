package haruka

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"os/user"
)

type View interface {
	Render(data interface{}) (err error)
	Write(context *Context)
}

type JsonView struct {
	Data []byte
}

func (v *JsonView) Render(data interface{}) (err error) {
	user.
		v.Data, err = json.Marshal(data)
	return err
}

func (v *JsonView) Write(context *Context) {
	context.Writer.Header().Set("Content-Type", "application/json")
	context.Writer.Write(v.Data)
}

func (c *Context) JSON(data interface{}) (err error) {
	view := JsonView{}
	err = view.Render(data)
	if err != nil {
		return
	}
	view.Write(c)
	return nil
}

type XMLView struct {
	Data []byte
}

func (v *XMLView) Render(data interface{}) (err error) {
	v.Data, err = xml.Marshal(data)
	return err
}

func (v *XMLView) Write(context *Context) {
	context.Writer.Header().Set("Content-Type", "text/xml")
	context.Writer.Write(v.Data)
}

func (c *Context) XML(data interface{}) (err error) {
	view := XMLView{}
	err = view.Render(data)
	if err != nil {
		return
	}
	view.Write(c)
	return nil
}

type HTMLView struct {
	Data     []byte
	Writer   io.Writer
	template *template.Template
}

func (v *HTMLView) Render(data interface{}) (err error) {
	return v.template.Execute(v.Writer, data)
}

func (v *HTMLView) Write(context *Context) {
	context.Writer.Header().Set("Content-Type", "text/html")
}

func (c *Context) HTML(templatePath string, data interface{}) (err error) {
	view := HTMLView{
		Writer:   c.Writer,
		template: template.Must(template.ParseFiles(templatePath)),
	}

	err = view.Render(data)
	if err != nil {
		return
	}
	view.Write(c)
	return nil
}

type JSON map[string]interface{}
type XML map[string]interface{}

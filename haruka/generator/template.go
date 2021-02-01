package generator

import (
	"os"
	"strings"
	"text/template"
)

type GoTemplate struct {
	path     string
	vars     map[string]interface{}
	output   string
	template *template.Template
}

func NewGoTemplate(output string) *GoTemplate {
	return &GoTemplate{
		output: output,
		vars:   map[string]interface{}{},
	}
}

func (t *GoTemplate) LoadTemplate(templateString string) *GoTemplate {
	templateString = strings.ReplaceAll(templateString, "$1", "`")
	t.template, _ = template.New("t1").Parse(templateString)
	return t
}
func (t *GoTemplate) AddVar(key string, data interface{}) *GoTemplate {
	t.vars[key] = data
	return t
}

func (t *GoTemplate) AddVars(vars map[string]interface{}) *GoTemplate {
	for key, value := range vars {
		t.vars[key] = value
	}
	return t
}
func (t *GoTemplate) GenerateCode() (err error) {
	file, err := os.OpenFile(t.output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()
	err = t.template.Execute(file, t.vars)
	return
}

package generator

import (
	"fmt"
	"github.com/allentom/haruka/haruka/generator/templates"
	"os"
	"strings"
	"text/template"
)

func GenerateRestModel(modelName string) error {
	captionName := strings.Title(modelName)
	outputName := fmt.Sprintf("handler_%s.go", modelName)
	file, err := os.OpenFile(outputName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer file.Close()
	if err != nil {
		panic(err)
	}
	data := make(map[string]string)
	data["modelName"] = captionName
	tmp, _ := template.New("t1").Parse(templates.RestTemplate)
	err = tmp.Execute(file, data)
	if err != nil {
		panic(err)
	}
	return nil
}

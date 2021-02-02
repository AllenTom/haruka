package generator

import (
	"github.com/allentom/haruka/haruka/generator/templates"
	"os"
	"path/filepath"
	"strings"
)

type newProjectOption struct {
	Name string
}

func NewProject(option newProjectOption) error {
	packageName := strings.ToLower(option.Name)
	err := NewGoTemplate("./main.go").
		LoadTemplate(templates.MainTemplate).
		AddVar("packageName", packageName).
		GenerateCode()
	if err != nil {
		return err
	}

	// create application
	applicationPath := "./application"
	err = os.MkdirAll(applicationPath, os.ModePerm)
	if err != nil {
		return err
	}
	err = NewGoTemplate(filepath.Join(applicationPath, "instance.go")).
		LoadTemplate(templates.InstanceTemplate).
		GenerateCode()
	if err != nil {
		return err
	}
	err = NewGoTemplate(filepath.Join(applicationPath, "config.go")).
		LoadTemplate(templates.ConfigTemplate).
		GenerateCode()
	if err != nil {
		return err
	}
	err = NewGoTemplate(filepath.Join(applicationPath, "router.go")).
		LoadTemplate(templates.RouterTemplate).
		GenerateCode()
	if err != nil {
		return err
	}

	// make database
	databasePath := "./database"
	err = os.MkdirAll(databasePath, os.ModePerm)
	if err != nil {
		return err
	}
	err = NewGoTemplate(filepath.Join(databasePath, "connection.go")).
		LoadTemplate(templates.SQLiteDatabaseConnectionTemplate).
		GenerateCode()
	if err != nil {
		return err
	}
	return nil
}

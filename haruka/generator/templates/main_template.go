package templates

var MainTemplate = `package {{ .packageName }}

import (
	"log"
	"{{ .packageName }}/application"
)

func main() {
	err := application.ReadConfig()
	if err != nil {
		log.Fatal(err)
	}
	application.RunApiService()
}
`

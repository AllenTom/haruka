package templates

var InstanceTemplate = `package application
import (
	"github.com/allentom/haruka"
	"github.com/allentom/haruka/middleware"
)

func RunApiService() {
	engine := haruka.NewEngine()
	engine.UseMiddleware(middleware.NewLoggerMiddleware())
	SetRouter(engine)
	engine.RunAndListen(AppConfig.Addr)
}
`

package main

import (
	"os"

	"github.com/unluckythoughts/go-microservice"
	"github.com/unluckythoughts/go-microservice/tools/web"
)

func exampleMiddleware(r web.MiddlewareRequest) error {
	r.GetContext().Logger().Info("test log from middleware")
	_ = r.SetContextValue("example-key", "example-value")
	return nil
}

func exampleHandler(r web.Request) (interface{}, error) {
	val := r.GetContext().Value("example-key").(string)
	r.GetContext().Logger().Infof("test log from handler with value: %s", val)
	r.GetContext().Logger().Errorf("test log from handler with value: %s", val)

	return "example-result", nil
}

func main() {
	_ = os.Setenv("DB_USER", "example")
	_ = os.Setenv("DB_PASSWORD", "example")
	_ = os.Setenv("DB_NAME", "example")
	_ = os.Setenv("WEB_PORT", "5679")
	_ = os.Setenv("WEB_CORS", "true")

	opts := microservice.Options{
		Name:        "example",
		EnableDB:    true,
		EnableCache: true,
		EnableBus:   true,
	}
	s := microservice.New(opts)

	s.HttpRouter().GET("/example", exampleMiddleware, exampleHandler)
	s.Start()
}

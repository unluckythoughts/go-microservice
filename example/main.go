package main

import (
	"os"

	"github.com/investing-bot/microservice"
	"github.com/investing-bot/microservice/tools/web"
)

func main() {
	os.Setenv("DB_USER", "example")
	os.Setenv("DB_PASSWORD", "example")
	os.Setenv("DB_NAME", "example")

	opts := microservice.Options{
		Name:        "example",
		EnableDB:    true,
		EnableCache: true,
		EnableBus:   true,
	}
	s := microservice.New(opts)
	s.HttpRouter().GET("/example", web.NotImplemented)
	s.Start()
}

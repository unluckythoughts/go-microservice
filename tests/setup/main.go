package main

import (
	"os"

	"github.com/unluckythoughts/go-microservice/v2"
)

func setupEnv() {
	_ = os.Setenv("DB_USER", "example")
	_ = os.Setenv("DB_PASSWORD", "example")
	_ = os.Setenv("DB_NAME", "example")
	_ = os.Setenv("WEB_PORT", "5679")
	_ = os.Setenv("WEB_CORS", "true")
}

func main() {
	setupEnv()
	opts := microservice.Options{
		Name:        "test-service",
		EnableDB:    true,
		EnableCache: true,
		EnableBus:   true,
		DBType:      microservice.DBTypePostgresql,
	}
	s := microservice.New(opts)
	registerRoutes(s.HttpRouter(), s.GetDB(), s.GetLogger())

	s.Start()
}

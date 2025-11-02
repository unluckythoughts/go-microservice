package main

import (
	"example/api"
	"example/service"

	"github.com/unluckythoughts/go-microservice"
)

func main() {
	opts := microservice.Options{
		Name:        "example-app",
		EnableDB:    true,
		DBType:      "postgresql",
		EnableCache: false,
		EnableBus:   false,
	}

	s := microservice.New(opts)
	serviceLayer := service.NewService(s.GetDB())

	// Register routes using the router from the server
	api.RegisterRoutes(s.HttpRouter(), serviceLayer)

	// Register socket methods (optional - only if sockets are enabled)
	// api.AddSocketHandlers(s, serviceLayer)

	s.Start()
}

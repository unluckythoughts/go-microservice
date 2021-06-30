package main

import (
	"os"

	"github.com/investing-bot/microservice"
)

func main() {
	os.Setenv("DB_USER", "example")
	os.Setenv("DB_PASSWORD", "example")
	os.Setenv("DB_NAME", "example")
	s := microservice.New("example")
	s.Start()
}

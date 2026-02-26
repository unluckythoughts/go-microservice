package main

import (
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func registerWebRoutes(r web.Router) {
	r.GET("/hello", func(ctx web.Context) any {
		return "Hello, World!"
	})
}

func registerRoutes(r web.Router, db *gorm.DB, l *zap.Logger) {
	as := auth.New(auth.Options{
		DB:     db,
		Logger: l,
	})

	err := auth.RegisterAuthRoutes(r, "/api/v1", as, 0)
	if err != nil {
		l.Fatal("Failed to register auth routes", zap.Error(err))
	}

	registerWebRoutes(r)
}

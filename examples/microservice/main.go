package main

import (
	"embed"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/unluckythoughts/go-microservice/v2"
	"github.com/unluckythoughts/go-microservice/v2/tools/auth"
	"github.com/unluckythoughts/go-microservice/v2/tools/bus"
	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrations embed.FS

func exampleMiddleware(r web.MiddlewareRequest) error {
	r.GetContext().Logger().Info("test log from middleware")

	return r.SetContextValue("example-key", "example-value")
}

func exampleHandler(r web.Request) (any, error) {
	val := r.GetContext().Value("example-key").(string)
	r.GetContext().Sugar().Infof("test log from handler with value: %s", val)
	r.GetContext().Sugar().Errorf("test log from handler with value: %s", val)

	return val, nil
}

func publishHandler(b bus.IBus) web.Handler {
	return func(r web.Request) (any, error) {
		err := b.Publish(bus.Message{
			Type:        "com-example-event",
			RoutingKeys: []string{"com.example.event"},
			Body:        []byte("hello world"),
		})
		if err != nil {
			return nil, errors.New("failed to publish message")
		}

		return "message published", nil
	}
}

const (
	UserRole  auth.Role = 1
	AdminRole auth.Role = 99
)

func runMigrations(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	src, err := iofs.New(migrations, "migrations")
	if err != nil {
		return err
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func main() {
	opts := microservice.Options{
		Name:        "example",
		EnableDB:    true,
		EnableCache: true, // enables Redis cache
		EnableBus:   true, // enables AWS SQS bus
	}
	s := microservice.New(opts)
	db := s.GetDB()
	if err := runMigrations(db); err != nil {
		panic(err)
	}

	as := auth.New(auth.Options{
		DB:     db,
		Logger: s.GetLogger().Named("auth"),
		UserRoles: map[auth.Role]string{
			UserRole:  "user",
			AdminRole: "admin",
		},
	})

	r := s.HttpRouter()
	auth.RegisterAuthRoutes(r, "/api/v1", as, UserRole)
	r.GET("/api/v1/example", exampleMiddleware, exampleHandler)

	b := s.GetBus()

	b.AddHandler("com.example.*", func(m bus.Message) error {
		s.GetLogger().Sugar().Infof("Received message: %s", string(m.Body))
		return nil
	})

	r.GET("/api/v1/bus/publish", publishHandler(b))

	s.Start()
}

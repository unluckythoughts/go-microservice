package microservice

import (
	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v8"
	"github.com/investing-bot/microservice/tools/bus"
	"github.com/investing-bot/microservice/tools/cache"
	"github.com/investing-bot/microservice/tools/logger"
	"github.com/investing-bot/microservice/tools/psql"
	"github.com/investing-bot/microservice/tools/sockets"
	"github.com/investing-bot/microservice/tools/web"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type (
	IService interface {
		Start()
		HttpRouter() web.Router
		SocketRegister(method string, handler sockets.Handler)
		GetDB() *gorm.DB
		GetCache() *redis.Client
		GetBus() bus.IBus
	}

	service struct {
		db     *gorm.DB
		cache  *redis.Client
		server *web.Server
		bus    bus.IBus
	}
)

func parseEnvironmentVars(config interface{}) {
	if err := env.Parse(config); err != nil {
		panic(errors.Wrapf(err, "could not parse env variables"))
	}
}

func getLogger() *zap.Logger {
	opts := logger.Options{}
	parseEnvironmentVars(&opts)
	return logger.New(opts)
}

func getServer(l *zap.Logger) *web.Server {
	opts := web.Options{}
	parseEnvironmentVars(&opts)
	opts.Logger = l

	return web.NewServer(opts)
}

func getDB(l *zap.Logger) *gorm.DB {
	opts := psql.Options{}
	parseEnvironmentVars(&opts)
	opts.Logger = l

	return psql.New(opts)
}

func getCache(l *zap.Logger) *redis.Client {
	opts := cache.Options{}
	parseEnvironmentVars(&opts)
	opts.Logger = l

	return cache.New(opts)
}

func getBus(l *zap.Logger) bus.IBus {
	opts := bus.Options{}
	parseEnvironmentVars(&opts)
	opts.Logger = l

	return bus.New(opts)
}

func New(name string) IService {
	l := getLogger()
	l.Named(name).Info("Starting " + name + " serice")
	s := getServer(l.Named(name))
	db := getDB(l.Named(name + ":db"))
	c := getCache(l.Named(name + ":cache"))
	b := getBus(l.Named(name + ":queue"))

	return &service{server: s, db: db, cache: c, bus: b}
}

func (s *service) Start() {
	s.server.Start()
}

func (s *service) HttpRouter() web.Router {
	return s.server.GetRouter()
}

func (s *service) SocketRegister(method string, handler sockets.Handler) {
	s.server.AddSocketHandler(method, handler)
}

func (s *service) GetDB() *gorm.DB {
	return s.db
}

func (s *service) GetCache() *redis.Client {
	return s.cache
}

func (s *service) GetBus() bus.IBus {
	return s.bus
}

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

	Options struct {
		Name        string `env:"SERVICE_NAME" envDefault:"true"`
		EnableDB    bool   `env:"SERVICE_ENABLE_DB" envDefault:"true"`
		EnableCache bool   `env:"SERVICE_ENABLE_CACHE" envDefault:"false"`
		EnableBus   bool   `env:"SERVICE_ENABLE_BUS" envDefault:"false"`
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

func New(opts Options) IService {
	l := getLogger()
	l.Named(opts.Name).Info("Starting " + opts.Name + " serice")
	s := &service{server: getServer(l.Named(opts.Name))}
	if opts.EnableDB {
		db := getDB(l.Named(opts.Name + ":db"))
		s.db = db
	}

	if opts.EnableBus {
		b := getBus(l.Named(opts.Name + ":queue"))
		s.bus = b
	}

	if opts.EnableCache {
		c := getCache(l.Named(opts.Name + ":cache"))
		s.cache = c
	}

	return s
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
	if s.db != nil {
		return s.db
	}

	panic("database is not configured with the service")
}

func (s *service) GetCache() *redis.Client {
	if s.cache != nil {
		return s.cache
	}

	panic("database is not configured with the service")
}

func (s *service) GetBus() bus.IBus {
	if s.bus != nil {
		return s.bus
	}

	panic("database is not configured with the service")
}

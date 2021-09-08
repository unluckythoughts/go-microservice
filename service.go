package microservice

import (
	"github.com/caarlos0/env"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"github.com/unluckythoughts/go-microservice/tools/bus"
	"github.com/unluckythoughts/go-microservice/tools/cache"
	"github.com/unluckythoughts/go-microservice/tools/logger"
	"github.com/unluckythoughts/go-microservice/tools/psql"
	"github.com/unluckythoughts/go-microservice/tools/sockets"
	"github.com/unluckythoughts/go-microservice/tools/web"
	"github.com/unluckythoughts/go-microservice/utils/alerts"
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
		GetAlerts() (*alerts.SlackClient, *alerts.TextClient)
		GetLogger() *zap.Logger
	}

	Options struct {
		Name        string `env:"SERVICE_NAME" envDefault:"true"`
		EnableDB    bool   `env:"SERVICE_ENABLE_DB" envDefault:"true"`
		EnableCache bool   `env:"SERVICE_ENABLE_CACHE" envDefault:"false"`
		EnableBus   bool   `env:"SERVICE_ENABLE_BUS" envDefault:"false"`
	}

	service struct {
		l      *zap.Logger
		db     *gorm.DB
		cache  *redis.Client
		server *web.Server
		bus    bus.IBus
		slack  *alerts.SlackClient
		text   *alerts.TextClient
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
	s := &service{l: l, server: getServer(l.Named(opts.Name))}

	s.slack = alerts.NewSlackClient(l.Named(opts.Name + ":slack"))
	s.text = alerts.NewTextClient(l.Named(opts.Name + ":text"))

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

func (s *service) GetAlerts() (*alerts.SlackClient, *alerts.TextClient) {
	return s.slack, s.text
}
func (s *service) GetLogger() *zap.Logger {
	return s.l
}

package microservice

import (
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/unluckythoughts/go-microservice/tools/bus"
	"github.com/unluckythoughts/go-microservice/tools/cache"
	"github.com/unluckythoughts/go-microservice/tools/logger"
	"github.com/unluckythoughts/go-microservice/tools/psql"
	"github.com/unluckythoughts/go-microservice/tools/sockets"
	"github.com/unluckythoughts/go-microservice/tools/sqlite"
	"github.com/unluckythoughts/go-microservice/tools/web"
	"github.com/unluckythoughts/go-microservice/utils"
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
		Name           string `env:"SERVICE_NAME" envDefault:"true"`
		EnableDB       bool   `env:"SERVICE_ENABLE_DB" envDefault:"true"`
		DBType         string `env:"SERVICE_DB_TYPE" envDefault:"postgresql"`
		EnableCache    bool   `env:"SERVICE_ENABLE_CACHE" envDefault:"false"`
		EnableBus      bool   `env:"SERVICE_ENABLE_BUS" envDefault:"false"`
		ProxyTransport web.ProxyTransport
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

const (
	DBTypePostgresql = "postgresql"
	DBTypeSqlite     = "sqlite"
)

func getLogger() *zap.Logger {
	opts := logger.Options{}
	utils.ParseEnvironmentVars(&opts)
	return logger.New(opts)
}

func getServer(l *zap.Logger) *web.Server {
	opts := web.Options{}
	utils.ParseEnvironmentVars(&opts)
	opts.Logger = l

	return web.NewServer(opts)
}

func getPsqlDB(l *zap.Logger) *gorm.DB {
	opts := psql.Options{}
	utils.ParseEnvironmentVars(&opts)
	opts.Logger = l

	return psql.New(opts)
}

func getSqliteDB(l *zap.Logger) *gorm.DB {
	opts := sqlite.Options{}
	utils.ParseEnvironmentVars(&opts)
	opts.Logger = l

	return sqlite.New(opts)
}

func getCache(l *zap.Logger) *redis.Client {
	opts := cache.Options{}
	utils.ParseEnvironmentVars(&opts)
	opts.Logger = l

	return cache.New(opts)
}

func getBus(l *zap.Logger) bus.IBus {
	opts := bus.Options{}
	utils.ParseEnvironmentVars(&opts)
	opts.Logger = l

	return bus.New(opts)
}

func New(opts Options) IService {
	logName := strings.ToLower(opts.Name)
	logName = strings.ReplaceAll(logName, " ", "-")
	l := getLogger().Named(logName)
	l.Info("Starting " + opts.Name + " service")
	s := &service{l: l, server: getServer(l.Named("web"))}

	if opts.ProxyTransport != nil {
		s.server.SetProxyTransport(opts.ProxyTransport)
	}

	s.slack = alerts.NewSlackClient(l.Named("slack"))
	s.text = alerts.NewTextClient(l.Named("text"))

	if opts.EnableDB {
		if opts.DBType == DBTypeSqlite {
			db := getSqliteDB(l.Named("db"))
			s.db = db
		} else {
			db := getPsqlDB(l.Named("db"))
			s.db = db
		}
	}

	if opts.EnableBus {
		b := getBus(l.Named("queue"))
		s.bus = b
	}

	if opts.EnableCache {
		c := getCache(l.Named("cache"))
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

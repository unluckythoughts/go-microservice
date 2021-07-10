package cache

import (
	"context"
	"crypto/tls"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type (
	Options struct {
		Logger     *zap.Logger
		Host       string `env:"CACHE_HOST" envDefault:"localhost"`
		Port       int    `env:"CACHE_PORT" envDefault:"6379"`
		Password   string `env:"CACHE_PASSWORD" envDefault:""`
		DB         int    `env:"CACHE_DB" envDefault:"0"`
		DisableSSL bool   `env:"CACHE_DISABLE_SSL" envDefault:"true"`
		Debug      bool   `env:"CACHE_DEBUG" envDefault:"false"`
	}
)

func sanityCheck(r *redis.Client) {
	if _, err := r.Ping(context.Background()).Result(); err != nil {
		panic(errors.Wrap(err, "Cache didn't respond to ping"))
	}

	if err := r.Set(context.Background(), "key", "value", time.Second).Err(); err != nil {
		panic(errors.Wrap(err, "Cache failed to connect"))
	}

	if _, err := r.Get(context.Background(), "key").Result(); err != nil {
		panic(errors.Wrap(err, "Cache failed to read"))
	}
}

func New(opts Options) *redis.Client {
	config := &redis.Options{
		Addr:     opts.Host + ":" + strconv.Itoa(opts.Port),
		Password: opts.Password,
		DB:       opts.DB,
	}

	if !opts.DisableSSL {
		config.TLSConfig = &tls.Config{
			// nolint:gosec
			InsecureSkipVerify: true,
		}
	}

	l := &cacheLogger{opts.Logger}

	redis.SetLogger(l)
	r := redis.NewClient(config)
	if opts.Debug {
		r.AddHook(l)
	}

	sanityCheck(r)
	opts.Logger.Info("Connected to redis")
	return r
}

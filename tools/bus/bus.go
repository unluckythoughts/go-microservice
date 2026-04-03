package bus

import (
	"fmt"

	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type Type uint

const (
	RabbitMQ Type = 1
	Kafka    Type = 2
)

type Options struct {
	Type   Type `env:"BUS_TYPE" envDefault:"1"`
	Logger *zap.Logger
	// RabbitMQ options
	Host     string `env:"BUS_HOST" envDefault:"localhost"`
	Port     int    `env:"BUS_PORT" envDefault:"5672"`
	User     string `env:"BUS_USER" envDefault:"guest"`
	Password string `env:"BUS_PASSWORD" envDefault:"guest"`
	AppName  string `env:"BUS_APP_NAME" envDefault:"api"`
}

type Handler func(msg Message) error

type IBus interface {
	AddHandler(queueName, topic string, handler Handler) (err error)
	Publish(msg Message) (err error)
}

type bus struct {
	l       *zap.SugaredLogger
	mq      *rabbitmq.Conn
	appName string
}

func getRabbitMQURL(opts Options) string {
	url := "amqp://"
	if opts.User != "" {
		url += opts.User
		if opts.Password != "" {
			url += ":" + opts.Password
		}
		url += "@"
	}
	url += opts.Host
	if opts.Port != 0 {
		url += ":" + fmt.Sprintf("%d", opts.Port)
	}
	url += "/"
	return url
}

func mqSanityCheck(conn *rabbitmq.Conn) error {
	if conn == nil {
		return fmt.Errorf("rabbitmq connection is nil")
	}
	return nil
}

func New(opts Options) IBus {
	switch opts.Type {
	case RabbitMQ:
		conn, err := rabbitmq.NewConn(
			getRabbitMQURL(opts),
			rabbitmq.WithConnectionOptionsLogger(opts.Logger.Sugar()),
		)
		if err != nil {
			panic(fmt.Errorf("could not connect to bus: %w", err))
		}
		if err := mqSanityCheck(conn); err != nil {
			panic(fmt.Errorf("rabbitmq sanity check failed: %w", err))
		}

		opts.Logger.Sugar().Infof("Connected to RabbitMQ at %s:%d", opts.Host, opts.Port)
		return &bus{l: opts.Logger.Sugar(), mq: conn, appName: opts.AppName}
	default:
		panic(fmt.Errorf("unsupported bus type: %d", opts.Type))
	}
}

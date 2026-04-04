package bus

import (
	"fmt"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/wagslane/go-rabbitmq"
	"go.uber.org/zap"
)

type Type uint

const (
	RabbitMQ Type = 1
	Kafka    Type = 2
)

type Options struct {
	Type   Type `env:"BUS_TYPE" envDefault:"2"`
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
	AddHandler(topic string, handler Handler) (err error)
	Publish(msg Message) (err error)
}

type bus struct {
	l       *zap.SugaredLogger
	appName string

	mqc *rabbitmq.Conn
	mqp *rabbitmq.Publisher
	kp  *kafka.Producer
	kc  *kafka.Consumer

	mut      sync.RWMutex
	once     sync.Once
	handlers map[string]Handler
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

func (b *bus) getMq(opts Options) (*rabbitmq.Conn, *rabbitmq.Publisher) {
	conn, err := rabbitmq.NewConn(
		getRabbitMQURL(opts),
		rabbitmq.WithConnectionOptionsLogger(opts.Logger.Sugar()),
	)
	if err != nil {
		panic(fmt.Errorf("could not connect to rabbitmq: %w", err))
	}
	if err := mqSanityCheck(conn); err != nil {
		panic(fmt.Errorf("rabbitmq sanity check failed: %w", err))
	}

	p, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogger(b.l.Named("publisher")),
		rabbitmq.WithPublisherOptionsExchangeKind("topic"),
		rabbitmq.WithPublisherOptionsExchangeName("topic_exchange"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsConfirm,
	)
	if err != nil {
		panic(fmt.Errorf("could not create publisher: %w", err))
	}

	p.NotifyPublish(func(conf rabbitmq.Confirmation) {
		if conf.Ack {
			b.l.Debugf("message acknowledged, delivery tag: %d", conf.DeliveryTag)
		} else {
			b.l.Warnf("message not acknowledged, delivery tag: %d", conf.DeliveryTag)
		}
	})

	p.NotifyReturn(func(ret rabbitmq.Return) {
		b.l.Errorf("message id returned: %s, reply code: %d, reply text: %s, exchange: %s, routing key: %s",
			ret.MessageId, ret.ReplyCode, ret.ReplyText, ret.Exchange, ret.RoutingKey)
	})
	return conn, p
}

func (b *bus) getKafka(opts Options) (*kafka.Producer, *kafka.Consumer) {
	kp, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":      fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		"client.id":              opts.AppName,
		"go.logs.channel.enable": true,
		"log_level":              7, // Debug level
	})
	if err != nil {
		panic(fmt.Errorf("could not create kafka producer: %w", err))
	}

	kc, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":                  fmt.Sprintf("%s:%d", opts.Host, opts.Port),
		"group.id":                           opts.AppName + "-group",
		"auto.offset.reset":                  "earliest",
		"allow.auto.create.topics":           true,
		"topic.metadata.refresh.interval.ms": 5000,
		"go.logs.channel.enable":             true,
		"log_level":                          7, // Debug level
	})
	if err != nil {
		panic(fmt.Errorf("could not create kafka consumer: %w", err))
	}

	return kp, kc
}

func New(opts Options) IBus {
	if opts.Type == 0 {
		opts.Type = Kafka
	}

	b := &bus{
		l:       opts.Logger.Sugar(),
		appName: opts.AppName,

		once: sync.Once{},
		mut:  sync.RWMutex{},
	}

	switch opts.Type {
	case RabbitMQ:
		conn, p := b.getMq(opts)
		opts.Logger.Sugar().Infof("Connected to RabbitMQ at %s:%d", opts.Host, opts.Port)
		b.mqc = conn
		b.mqp = p
		return b
	case Kafka:
		kp, kc := b.getKafka(opts)
		b.kp = kp
		b.kc = kc
		b.logKafkaMessages()
		opts.Logger.Sugar().Infof("Connected to Kafka at %s:%d", opts.Host, opts.Port)
		return b
	default:
		panic(fmt.Errorf("unsupported bus type: %d", opts.Type))
	}
}

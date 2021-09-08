package bus

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Options struct {
	Logger          *zap.Logger
	AccessKeyID     string `env:"AWS_ACCESS_KEY" envDefault:"NonEmptyLocalAccessKeyID"`
	SecretAccessKey string `env:"AWS_ACCESS_SECRET" envDefault:"NonEmptyLocalSecretAccessKey"`
	SessionToken    string `env:"AWS_ACCESS_TOKEN" envDefault:"NonEmptyLocalSessionToken"`
	Region          string `env:"AWS_REGION" envDefault:"us-west-2"`

	Environment string `env:"QUEUE_ENVIRONMENT" envDefault:"local"`
	EndpointURL string `env:"QUEUE_ENDPOINT_URL" envDefault:"http://localhost:4566"`
	DisableSSL  bool   `env:"QUEUE_DISABLE_SSL" envDefault:"true"`
	Debug       bool   `env:"QUEUE_DEBUG" envDefault:"false"`
}

const (
	EnvironmentLocal = "local"
)

type Handler func(msg Message) error

type IBus interface {
	AddHandler(queueName string, handler Handler) (err error)
	Publish(queueName string, msg Message) (err error)
}

type queueService struct {
	l   *zap.Logger
	sqs *sqs.SQS
}

func sanityCheck(q *sqs.SQS) {
	params := sqs.ListQueuesInput{}
	_, err := q.ListQueues(&params)
	if err != nil {
		panic(errors.Wrapf(err, "could not connect to sqs"))
	}
}

func New(opts Options) IBus {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	conf := aws.NewConfig()
	if opts.Environment == EnvironmentLocal {
		conf.WithEndpoint(opts.EndpointURL)
		conf.WithRegion(opts.Region)
		conf.WithCredentials(credentials.NewStaticCredentials(
			opts.AccessKeyID, opts.SecretAccessKey, opts.SessionToken))
	}
	conf.WithDisableSSL(opts.DisableSSL)
	conf.WithLogger(&queueLogger{opts.Logger})
	if opts.Debug {
		conf.WithLogLevel(aws.LogDebug)
	}

	q := sqs.New(sess, conf)

	sanityCheck(q)
	opts.Logger.Info("Connected to SQS")
	return &queueService{sqs: q, l: opts.Logger}
}

func (qs *queueService) getQueueURL(name string) (url *string, err error) {
	params := sqs.GetQueueUrlInput{
		QueueName: &name,
	}
	output, err := qs.sqs.GetQueueUrl(&params)
	if err != nil {
		return nil, err
	}

	return output.QueueUrl, nil
}

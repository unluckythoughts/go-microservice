package bus

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
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
	sqs *sqs.Client
}

func sanityCheck(q *sqs.Client) {
	params := sqs.ListQueuesInput{}
	_, err := q.ListQueues(context.Background(), &params)
	if err != nil {
		panic(errors.Wrapf(err, "could not connect to sqs"))
	}
}

func New(opts Options) IBus {
	ctx := context.Background()
	loadOptions := []func(*config.LoadOptions) error{
		config.WithRegion(opts.Region),
	}

	if opts.Environment == EnvironmentLocal {
		loadOptions = append(loadOptions,
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				opts.AccessKeyID, opts.SecretAccessKey, opts.SessionToken,
			)),
		)
	}

	cfg, err := config.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		panic(errors.Wrap(err, "could not load AWS config"))
	}

	q := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		if opts.Environment == EnvironmentLocal {
			o.BaseEndpoint = aws.String(opts.EndpointURL)
		}
	})

	sanityCheck(q)
	opts.Logger.Info("Connected to SQS")
	return &queueService{sqs: q, l: opts.Logger}
}

func (qs *queueService) getQueueURL(name string) (url *string, err error) {
	params := sqs.GetQueueUrlInput{
		QueueName: aws.String(name),
	}
	output, err := qs.sqs.GetQueueUrl(context.Background(), &params)
	if err != nil {
		return nil, err
	}

	return output.QueueUrl, nil
}

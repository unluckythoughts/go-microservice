package bus

import (
	"fmt"
	"sync"

	"github.com/wagslane/go-rabbitmq"
)

var (
	publisher    *rabbitmq.Publisher
	once         sync.Once
	publisherErr error
)

func (b *bus) initPublisher() error {
	once.Do(func() {
		p, err := rabbitmq.NewPublisher(
			b.mq,
			rabbitmq.WithPublisherOptionsLogger(b.l.Named("publisher")),
			rabbitmq.WithPublisherOptionsExchangeKind("topic"),
			rabbitmq.WithPublisherOptionsExchangeName("topic_exchange"),
			rabbitmq.WithPublisherOptionsExchangeDeclare,
			rabbitmq.WithPublisherOptionsConfirm,
		)
		if err != nil {
			publisherErr = fmt.Errorf("could not create publisher: %w", err)
			return
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

		publisher = p
	})

	return publisherErr
}

func (b *bus) Publish(msg Message) (err error) {
	if b.mq != nil {
		if err := b.initPublisher(); err != nil {
			return err
		}

		return publisher.Publish(
			msg.Body,
			msg.RoutingKeys,
			rabbitmq.WithPublishOptionsExchange("topic_exchange"),
			rabbitmq.WithPublishOptionsCorrelationID(msg.CorelationID),
			rabbitmq.WithPublishOptionsMessageID(msg.ID),
			rabbitmq.WithPublishOptionsType(msg.Type),
			rabbitmq.WithPublishOptionsAppID(b.appName),
		)
	}

	return fmt.Errorf("bus is not initialized")
}

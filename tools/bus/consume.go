package bus

import (
	"fmt"

	"github.com/wagslane/go-rabbitmq"
)

func (b *bus) handleMQMessage(handler Handler) func(d rabbitmq.Delivery) rabbitmq.Action {
	return func(d rabbitmq.Delivery) rabbitmq.Action {
		msg := Message{
			ID:           d.MessageId,
			CorelationID: d.CorrelationId,
			Type:         d.Type,
			PublishTime:  d.Timestamp,
			Body:         d.Body,
		}

		if err := handler(msg); err != nil {
			b.l.Errorf("error handling message: %v", err)
			return rabbitmq.NackRequeue
		}

		return rabbitmq.Ack
	}
}

func (b *bus) addMqHandler(queueName, topic string, handler Handler) error {
	consumer, err := rabbitmq.NewConsumer(
		b.mq,
		queueName,
		rabbitmq.WithConsumerOptionsLogger(b.l.Named(queueName)),
		rabbitmq.WithConsumerOptionsExchangeKind("topic"),
		rabbitmq.WithConsumerOptionsExchangeName("topic_exchange"),
		rabbitmq.WithConsumerOptionsRoutingKey(topic),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		return err
	}

	go func() {
		err = consumer.Run(b.handleMQMessage(handler))
		if err != nil {
			b.l.Errorf("Bus consumer of %s stopped with error: %v", queueName, err)
			consumer.Close()
		}
	}()

	return nil
}

func (b *bus) AddHandler(queueName, topic string, handler Handler) (err error) {
	if b.mq != nil {
		return b.addMqHandler(queueName, topic, handler)
	}

	return fmt.Errorf("no message bus configured")
}

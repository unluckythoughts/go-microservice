package bus

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
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

func (b *bus) handleKafkaMessages() {
	b.once.Do(func() {
		go func() {
			for {
				msg, err := b.kc.ReadMessage(-1)
				if err != nil {
					if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.Code() == kafka.ErrUnknownTopicOrPart {
						continue // transient: no topics matching regex exist yet
					}
					b.l.Errorf("error reading message from Kafka: %v", err)
					continue
				}

				handler, exists := b.getHandler(*msg.TopicPartition.Topic)
				if !exists {
					b.l.Warnf("no handler found for topic: %s", *msg.TopicPartition.Topic)
					continue
				}

				var m Message
				if err := json.Unmarshal(msg.Value, &m); err != nil {
					b.l.Errorf("error unmarshalling message: %v", err)
					continue
				}

				if err := handler(m); err != nil {
					b.l.Errorf("error handling message: %v", err)
				}
			}
		}()
	})
}

func (b *bus) addMqHandler(topic string, handler Handler) error {
	consumer, err := rabbitmq.NewConsumer(
		b.mqc,
		topic+"_queue",
		rabbitmq.WithConsumerOptionsLogger(b.l.Named(topic+"_queue")),
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
			b.l.Errorf("Bus consumer of %s stopped with error: %v", topic+"_queue", err)
			consumer.Close()
		}
	}()

	return nil
}

func (b *bus) addKafkaHandler(topic string, _ Handler) error {
	topics, err := b.kc.Subscription()
	if err != nil {
		return fmt.Errorf("failed to get Kafka subscriptions: %w", err)
	}

	found := false
	for _, t := range topics {
		if t == topic {
			found = true
			break
		}
	}

	if found {
		return fmt.Errorf("topic %s already subscribed in Kafka", topic)
	}

	topic = strings.ReplaceAll(topic, "#", ".*")
	topic = strings.ReplaceAll(topic, "*", "[^.]*")
	topic = "^" + topic + "$"

	topics = append(topics, topic)

	err = b.kc.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, err)
	}

	b.handleKafkaMessages()

	return nil
}

func (b *bus) updateHandlers(topic string, handler Handler) {
	b.mut.Lock()
	defer b.mut.Unlock()

	if b.handlers == nil {
		b.handlers = make(map[string]Handler)
	}
	b.handlers[topic] = handler
}

func (b *bus) getHandler(topic string) (Handler, bool) {
	b.mut.RLock()
	defer b.mut.RUnlock()

	handler, exists := b.handlers[topic]
	if exists {
		return handler, true
	}

	// If no exact match, check for pattern matches for kafka topics
	for pattern, handler := range b.handlers {
		regexPattern := strings.ReplaceAll(pattern, "#", ".*")
		regexPattern = strings.ReplaceAll(regexPattern, "*", "[^.]*")
		regexPattern = "^" + regexPattern + "$"
		if matched, _ := regexp.MatchString(regexPattern, topic); matched {
			return handler, true
		}
	}
	return nil, false
}

func (b *bus) AddHandler(topic string, handler Handler) (err error) {
	b.updateHandlers(topic, handler)

	if b.mqc != nil {
		return b.addMqHandler(topic, handler)
	}

	if b.kc != nil {
		return b.addKafkaHandler(topic, handler)
	}

	return fmt.Errorf("no message bus configured")
}

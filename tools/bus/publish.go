package bus

import (
	"encoding/json"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/wagslane/go-rabbitmq"
)

func (b *bus) Publish(msg Message) (err error) {
	if b.mqc != nil {
		return b.mqp.Publish(
			msg.Body,
			msg.RoutingKeys,
			rabbitmq.WithPublishOptionsExchange("topic_exchange"),
			rabbitmq.WithPublishOptionsCorrelationID(msg.CorelationID),
			rabbitmq.WithPublishOptionsMessageID(msg.ID),
			rabbitmq.WithPublishOptionsType(msg.Type),
			rabbitmq.WithPublishOptionsAppID(b.appName),
		)
	}

	if b.kp != nil {
		msgBytes, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("could not marshal message: %w", err)
		}

		for _, topic := range msg.RoutingKeys {
			err := b.kp.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          msgBytes,
			}, nil)
			if err != nil {
				return fmt.Errorf("could not produce message to Kafka: %w", err)
			}
		}
		return nil
	}

	return fmt.Errorf("bus is not initialized")
}

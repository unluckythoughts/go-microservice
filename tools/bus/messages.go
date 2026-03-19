package bus

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Message struct {
	ID     string
	Body   string
	Logger *zap.Logger
}

func (qs *queueService) deleteMessage(url *string, m types.Message) (err error) {
	params := sqs.DeleteMessageInput{
		ReceiptHandle: m.ReceiptHandle,
		QueueUrl:      url,
	}
	_, err = qs.sqs.DeleteMessage(context.Background(), &params)
	if err != nil {
		return errors.Wrapf(err, "could not delete message with id: %s", aws.ToString(m.MessageId))
	}

	return nil
}

func (qs *queueService) Publish(queueName string, msg Message) (err error) {
	url, err := qs.getQueueURL(queueName)
	if err != nil {
		return errors.Wrapf(err, "could not get queue %s url", queueName)
	}

	params := sqs.SendMessageInput{
		MessageBody: aws.String(msg.Body),
		QueueUrl:    url,
	}

	if _, err := qs.sqs.SendMessage(context.Background(), &params); err != nil {
		return errors.Wrapf(err, "could not publish message onto queue %s", queueName)
	}

	return nil
}

func (qs *queueService) AddHandler(queueName string, handler Handler) (err error) {
	url, err := qs.getQueueURL(queueName)
	if err != nil {
		return errors.Wrapf(err, "could not get the queue url")
	}

	go func() {
		params := sqs.ReceiveMessageInput{
			QueueUrl:        url,
			WaitTimeSeconds: 20,
		}

		for {
			resp, err := qs.sqs.ReceiveMessage(context.Background(), &params)
			if err != nil {
				qs.l.Sugar().Errorf("could not receive messages from queue %s. error: %+v", queueName, err)
			}

			for _, m := range resp.Messages {
				l := qs.l.With(
					zap.String("queue", queueName),
					zap.String("messageId", aws.ToString(m.MessageId)),
				)
				msg := Message{
					ID:     aws.ToString(m.MessageId),
					Body:   aws.ToString(m.Body),
					Logger: l,
				}

				if err := handler(msg); err != nil {
					l.Error("could not process message", zap.Error(err))
				} else {
					if err := qs.deleteMessage(url, m); err != nil {
						l.Error("error deleting message", zap.Error(err))
					}
				}
			}
		}
	}()

	return nil
}

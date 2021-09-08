package bus

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Message struct {
	ID     string
	Body   string
	Logger *zap.Logger
}

func (qs *queueService) deleteMessage(url *string, m *sqs.Message) (err error) {
	params := sqs.DeleteMessageInput{
		ReceiptHandle: m.ReceiptHandle,
		QueueUrl:      url,
	}
	_, err = qs.sqs.DeleteMessage(&params)
	if err != nil {
		return errors.Wrapf(err, "could not delete message with id: %s", *m.MessageId)
	}

	return nil
}

func (qs *queueService) Publish(queueName string, msg Message) (err error) {
	url, err := qs.getQueueURL(queueName)
	if err != nil {
		return errors.Wrapf(err, "could not get queue %s url", queueName)
	}

	params := sqs.SendMessageInput{
		MessageBody: &msg.Body,
		QueueUrl:    url,
	}

	if _, err := qs.sqs.SendMessage(&params); err != nil {
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
			WaitTimeSeconds: aws.Int64(20),
		}

		for {
			resp, err := qs.sqs.ReceiveMessage(&params)
			if err != nil {
				qs.l.Sugar().Errorf("could not receive messages from queue %s. error: %+v", queueName, err)
			}

			for _, m := range resp.Messages {
				l := qs.l.With(
					zap.String("queue", queueName),
					zap.String("messageId", *m.MessageId),
				)
				msg := Message{
					ID:     *m.MessageId,
					Body:   *m.Body,
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

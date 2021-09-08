package alerts

import (
	"net/http"

	"github.com/unluckythoughts/go-microservice/tools/web"
	"go.uber.org/zap"
)

type SlackClient struct {
	webClient web.Client
	l         *zap.SugaredLogger
}

const (
	// TODO: Slack channel
	baseSlackURL = ""
)

func NewSlackClient(l *zap.Logger) *SlackClient {
	defaultHeaders := http.Header{"Content-Type": []string{"application/json"}}
	return &SlackClient{
		webClient: web.NewClient(baseSlackURL, defaultHeaders),
		l:         l.Sugar(),
	}
}

func (c *SlackClient) SendMessage(message, channelURL string) {
	body := struct {
		Text string `json:"text"`
	}{Text: message}

	status, err := c.webClient.PostResponse(channelURL, body, nil)
	if err != nil {
		c.l.Errorf("Error posting message %s to slack, error: %s", message, err.Error())
	} else if status != http.StatusOK {
		c.l.Info("Received status %d while posting message %s to slack", status, message)
	}
}

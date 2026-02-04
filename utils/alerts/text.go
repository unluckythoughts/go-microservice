package alerts

import (
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/unluckythoughts/go-microservice/v2/tools/web"
	"go.uber.org/zap"
)

type TextClient struct {
	webClient web.Client
	l         *zap.SugaredLogger
}

const (
	// TODO: twilio tokens
	accountSID    = ""
	authToken     = ""
	baseTwilioURL = "https://api.twilio.com/2010-04-01/Accounts/" + accountSID

	// TODO: twilio from number
	fromNumber  = ""
	messagesUrl = "/Messages.json"
)

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func NewTextClient(l *zap.Logger) *TextClient {
	defaultHeaders := make(http.Header)
	defaultHeaders.Add("Accept", "application/json")
	defaultHeaders.Add("Content-Type", "application/x-www-form-urlencoded")
	defaultHeaders.Set("Authorization", basicAuth(accountSID, authToken))

	return &TextClient{
		webClient: web.NewClient(baseTwilioURL, defaultHeaders),
		l:         l.Sugar(),
	}
}

func getRequestBody(message, number string) []byte {
	msgData := url.Values{}
	msgData.Set("To", number)
	msgData.Set("From", fromNumber)
	msgData.Set("Body", message)
	return []byte(msgData.Encode())
}

func (c *TextClient) SendMessage(message, number string) {
	body := getRequestBody(message, number)
	resp := map[string]interface{}{}

	status, err := c.webClient.PostResponse(messagesUrl, body, &resp)
	if err != nil {
		c.l.Errorf("Error sending text message %s to %s, error: %s", message, number, err.Error())
	} else if status != http.StatusOK {
		c.l.Info("Received status %d while sending text message %s to %s, response: %+v", status, message, number, resp)
	}
}

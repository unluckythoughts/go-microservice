package main

import (
	"github.com/unluckythoughts/go-microservice/v2/integrations/text"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"github.com/unluckythoughts/go-microservice/v2/tools/logger"
)

func main() {
	l := logger.New(logger.Options{})
	s := text.New(&text.Options{
		Token:    "<your-magictext-token>",
		SenderID: "<your-magictext-sender-id>",
	})

	s.MagicTextSMS(context.NewContext(l), &text.Message{
		To:         []string{"<recipient-phone-number>"},
		TemplateID: "<your-magictext-template-id>",
		Body:       "<your-message-body-according-to-template>",
	})
}

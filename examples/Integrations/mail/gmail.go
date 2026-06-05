package main

import (
	"fmt"

	"github.com/unluckythoughts/go-microservice/v2/integrations/mail"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"github.com/unluckythoughts/go-microservice/v2/tools/logger"
)

func main() {
	s, err := mail.New(&mail.Options{
		Host:        "smtp.gmail.com",
		Port:        587,
		Encryption:  "tls",
		Username:    "<your-email@gmail.com>",
		AppPassword: "<your-app-password>",
	})
	if err != nil {
		panic(err)
	}

	err = s.SendEmail(context.NewContext(logger.New(logger.Options{})), &mail.Email{
		To:      []string{"gvsmraju@gmail.com"},
		Subject: "Test Email",
		Body:    "This is a test email sent using the Gmail integration.",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Email sent successfully!")
}

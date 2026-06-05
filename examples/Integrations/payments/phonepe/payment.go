package main

import (
	"strings"

	"github.com/gofrs/uuid"
	"github.com/unluckythoughts/go-microservice/v2/integrations/payments/phonepe"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"github.com/unluckythoughts/go-microservice/v2/tools/logger"
)

func main() {
	s := phonepe.New(phonepe.Options{
		Environment:  "sandbox",
		MerchantID:   "M23627PHURCSC",
		ClientID:     "M23627PHURCSC_2606031525",
		ClientSecret: "ZjMxNjA5MjEtNDZmYS00YzhkLTk2NTEtMTAwYTc5NGFlZTlj",
	})

	ctx := context.NewContext(logger.New(logger.Options{}))

	uuid, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	orderID := strings.ReplaceAll(uuid.String(), "-", "")
	resp, err := s.CreatePayment(ctx, &phonepe.PhonePePaymentRequest{
		MerchantOrderID: orderID,
		Amount:          10000,
		Description:     "Test payment",
	})
	if err != nil {
		panic(err)
	}
	println("Payment created with Order ID:", resp)

	status, err := s.GetPaymentStatus(ctx, orderID)
	if err != nil {
		panic(err)
	}
	println("Payment status:", status)

}

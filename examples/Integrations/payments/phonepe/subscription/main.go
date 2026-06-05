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

	subscriptionUUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	orderUUID, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	merchantSubscriptionID := strings.ReplaceAll(subscriptionUUID.String(), "-", "")
	merchantOrderID := strings.ReplaceAll(orderUUID.String(), "-", "")

	resp, err := s.CreateSubscription(ctx, &phonepe.PhonePeSubscriptionRequest{
		MerchantSubscriptionID: merchantSubscriptionID,
		MerchantOrderID:        merchantOrderID,
		Amount:                 100, // ₹1 — first transaction amount for TRANSACTION auth
		AuthWorkflowType:       "TRANSACTION",
		AmountType:             "FIXED",
		MaxAmount:              100, // ₹1 max per cycle
		Frequency:              "MONTHLY",
	})
	if err != nil {
		panic(err)
	}
	println("Subscription setup initiated:")
	println("  OrderID:", resp.OrderID)
	println("  State:", resp.State)
	println("  RedirectURL:", resp.RedirectURL)
	println("  IntentURL:", resp.IntentURL)
	println("  QrData:", resp.QrData)

	status, err := s.GetSubscriptionStatus(ctx, merchantSubscriptionID)
	if err != nil {
		panic(err)
	}
	println("Subscription status:", status)
}

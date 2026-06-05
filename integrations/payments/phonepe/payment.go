package phonepe

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	request "github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/models/request"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"go.uber.org/zap"
)

type PhonePePaymentRequest struct {
	// Unique identifier for the payment
	MerchantOrderID string `json:"merchantOrderId"`
	// Amount in paise (e.g., 10000 for ₹100)
	Amount int64 `json:"amount"`
	// Description of the payment
	Description string `json:"description"`
}

func (s *Service) CreatePayment(ctx context.Context, req *PhonePePaymentRequest) (string, error) {
	l := ctx.Logger()

	l.Debug("Creating a payment using PhonePe SDK")
	// Create a new payment request
	paymentRequest := request.StandardCheckoutPayRequest{
		MerchantOrderID: req.MerchantOrderID,
		Amount:          req.Amount,
		MetaInfo:        &models.MetaInfo{Udf1: req.Description},
		PaymentFlow:     request.NewPgCheckoutPaymentFlow(req.Description, nil, nil),
	}

	resp, err := s.c.Pay(ctx, &paymentRequest)
	if err != nil {
		l.Error("Failed to create payment", zap.Error(err))
		return "", err
	}

	l.Debug("Payment created successfully", zap.Any("response", resp))
	return resp.OrderID, nil
}

func (s *Service) GetPaymentStatus(ctx context.Context, merchantOrderID string) (any, error) {
	l := ctx.Logger()

	l.Debug("Fetching payment status using PhonePe SDK", zap.String("merchantOrderID", merchantOrderID))
	resp, err := s.c.GetOrderStatus(ctx, merchantOrderID, true)
	if err != nil {
		l.Error("Failed to fetch payment status", zap.Error(err))
		return nil, err
	}

	l.Debug("Payment status fetched successfully", zap.Any("response", resp))
	return resp, nil
}

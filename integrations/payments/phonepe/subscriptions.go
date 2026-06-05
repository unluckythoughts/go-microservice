package phonepe

import (
	"github.com/PhonePe/phonepe-pg-sdk-go/common/models"
	commonInstruments "github.com/PhonePe/phonepe-pg-sdk-go/common/models/instruments"
	commonRequest "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request"
	requestInstruments "github.com/PhonePe/phonepe-pg-sdk-go/common/models/request/instruments"
	subscriptionRequest "github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2/models/request"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"go.uber.org/zap"
)

type PhonePeSubscriptionRequest struct {
	// Unique identifier for the subscription
	MerchantSubscriptionID string `json:"merchantSubscriptionId"`
	// Unique identifier for the setup order
	MerchantOrderID string `json:"merchantOrderId"`
	// Amount in paise for the setup transaction (e.g., 100 for ₹1 PENNY_DROP)
	Amount int64 `json:"amount"`
	// AuthWorkflowType is the auth workflow. One of "PENNY_DROP" or "TRANSACTION"
	AuthWorkflowType string `json:"authWorkflowType"`
	// AmountType is the debit amount type. One of "FIXED" or "VARIABLE"
	AmountType string `json:"amountType"`
	// MaxAmount is the maximum debit amount per cycle in paise
	MaxAmount int64 `json:"maxAmount"`
	// Frequency is the subscription frequency. One of "DAILY", "WEEKLY", "MONTHLY", "YEARLY", etc.
	Frequency string `json:"frequency"`
	// ExpireAt is the optional Unix timestamp in milliseconds when the subscription expires
	ExpireAt *int64 `json:"expireAt,omitempty"`
	// VPA is the optional UPI VPA for UPI collect-based setup
	VPA string `json:"vpa,omitempty"`
}

type PhonePeSubscriptionResponse struct {
	// OrderID is the PhonePe order ID for the setup transaction
	OrderID string `json:"orderId"`
	// State is the current state of the order (e.g. "PENDING")
	State string `json:"state"`
	// RedirectURL is the web redirect URL to complete mandate setup (browser flow)
	RedirectURL string `json:"redirectUrl,omitempty"`
	// IntentURL is the UPI intent URL to launch the UPI app for mandate setup
	IntentURL string `json:"intentUrl,omitempty"`
	// QrData is the QR code data for mandate setup via QR scan
	QrData string `json:"qrData,omitempty"`
}

func (s *Service) CreateSubscription(ctx context.Context, req *PhonePeSubscriptionRequest) (*PhonePeSubscriptionResponse, error) {
	l := ctx.Logger()
	l.Debug("Creating a PhonePe subscription", zap.String("merchantSubscriptionId", req.MerchantSubscriptionID))

	var paymentMode commonInstruments.PaymentV2Instrument
	if req.VPA != "" {
		details := requestInstruments.NewVpaCollectPaymentDetails(req.VPA)
		paymentMode = commonInstruments.NewCollectPaymentV2Instrument(details, "")
	} else {
		paymentMode = commonInstruments.NewIntentPaymentV2InstrumentWithoutTargetApp()
	}

	paymentFlow := subscriptionRequest.NewSubscriptionSetupPaymentFlow(
		req.MerchantSubscriptionID,
		subscriptionRequest.AuthWorkflowType(req.AuthWorkflowType),
		subscriptionRequest.AmountType(req.AmountType),
		req.MaxAmount,
		subscriptionRequest.Frequency(req.Frequency),
		req.ExpireAt,
		paymentMode,
	)

	paymentRequest := &commonRequest.PgPaymentRequest{
		MerchantOrderID: req.MerchantOrderID,
		Amount:          req.Amount,
		MetaInfo:        models.MetaInfo{},
		PaymentFlow:     paymentFlow,
	}

	resp, err := s.sc.Setup(ctx, paymentRequest)
	if err != nil {
		l.Error("Failed to create PhonePe subscription", zap.Error(err))
		return nil, err
	}

	l.Debug("PhonePe subscription created successfully", zap.String("orderId", resp.OrderID))
	return &PhonePeSubscriptionResponse{
		OrderID:     resp.OrderID,
		State:       resp.State,
		RedirectURL: resp.RedirectURL,
		IntentURL:   resp.IntentURL,
		QrData:      resp.QrData,
	}, nil
}

func (s *Service) GetSubscriptionStatus(ctx context.Context, merchantSubscriptionID string) (any, error) {
	l := ctx.Logger()
	l.Debug("Fetching PhonePe subscription status", zap.String("merchantSubscriptionId", merchantSubscriptionID))

	resp, err := s.sc.GetSubscriptionStatus(ctx, merchantSubscriptionID)
	if err != nil {
		l.Error("Failed to fetch PhonePe subscription status", zap.Error(err))
		return nil, err
	}

	l.Debug("PhonePe subscription status fetched", zap.Any("response", resp))
	return resp, nil
}

func (s *Service) CancelSubscription(ctx context.Context, merchantSubscriptionID string) error {
	l := ctx.Logger()
	l.Debug("Cancelling PhonePe subscription", zap.String("merchantSubscriptionId", merchantSubscriptionID))

	err := s.sc.CancelSubscription(ctx, merchantSubscriptionID)
	if err != nil {
		l.Error("Failed to cancel PhonePe subscription", zap.Error(err))
		return err
	}

	l.Debug("PhonePe subscription cancelled successfully")
	return nil
}

package pan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"go.uber.org/zap"
)

type panRequest struct {
	PANNumber string `json:"pan"`
}

type panDetails struct {
	NameValidated  string `json:"name_validated"`
	Name           string `json:"name"`
	NameMatchScore int    `json:"name_match_score"`
	PAN            string `json:"pan"`
	SeedingStatus  string `json:"seeding_status"`
	NameMatch      bool   `json:"name_match"`
	PANDisplayName string `json:"pan_display_name"`
	Status         string `json:"status"`
}

type panQueryResponse struct {
	ResultCode       int         `json:"result_code"`
	ClientRefNum     string      `json:"client_ref_num"`
	RequestID        string      `json:"request_id"`
	HTTPResponseCode int         `json:"http_response_code"`
	Result           *panDetails `json:"result"`
}

type panResponse struct {
	Msg       string            `json:"msg"`
	Charges   float64           `json:"charges"`
	Data      *panQueryResponse `json:"data"`
	ClsBal    float64           `json:"clsBal"`
	RequestID string            `json:"requestID"`
	GST       float64           `json:"gst"`
	ErrorCode int               `json:"error_code"`
	Status    string            `json:"status"`
}

func (s *Service) IMWalletGetPANDetails(ctx context.Context, panNumber string) (PANDetails, error) {
	l := ctx.Logger()
	apiURL := "https://partner.imwallet.in/web_services/verificationSuit/walletBased/panBasic2.jsp"

	body := panRequest{
		PANNumber: panNumber,
	}

	data, err := json.Marshal(body)
	if err != nil {
		l.Error("Failed to marshal request body", zap.Error(err))
		return PANDetails{}, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(data))
	if err != nil {
		l.Error("Failed to create request", zap.Error(err))
		return PANDetails{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("userCode", s.code)
	req.Header.Set("webToken", s.token)
	resp, err := s.c.Do(req)
	if err != nil {
		l.Error("Failed to make request", zap.Error(err))
		return PANDetails{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.Error("Received non-200 response", zap.Int("status_code", resp.StatusCode))
		return PANDetails{}, fmt.Errorf("received non-200 response: %d", resp.StatusCode)
	}

	var respData panResponse
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		l.Error("Failed to decode response body", zap.Error(err))
		return PANDetails{}, err
	}

	if respData.Status != "success" {
		l.Error("API returned error status", zap.String("status", respData.Status), zap.Int("error_code", respData.ErrorCode), zap.String("message", respData.Msg))
		return PANDetails{}, fmt.Errorf("API error: %s (code: %d)", respData.Msg, respData.ErrorCode)
	}

	if respData.Data == nil || respData.Data.Result == nil {
		l.Error("API response missing data", zap.Any("response", respData))
		return PANDetails{}, fmt.Errorf("API response missing data")
	}

	if respData.Data.Result.PAN != panNumber {
		l.Error("API response PAN does not match request", zap.String("requested_pan", panNumber), zap.String("response_pan", respData.Data.Result.PAN))
		return PANDetails{}, fmt.Errorf("API response PAN does not match request")
	}

	if respData.Data.Result.NameValidated != "Y" {
		l.Error("Name validation failed for PAN", zap.String("pan", panNumber), zap.String("name_validated", respData.Data.Result.NameValidated))
		return PANDetails{}, fmt.Errorf("name validation failed for PAN: %s", panNumber)
	}

	result := respData.Data.Result
	panDetails := PANDetails{
		Number: result.PAN,
		Name:   result.Name,
		Status: result.Status,
	}

	l.Debug("Successfully fetched PAN details", zap.Any("pan_details", panDetails))
	return panDetails, nil
}

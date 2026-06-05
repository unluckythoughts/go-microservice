package text

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/unluckythoughts/go-microservice/v2/tools/context"
	"go.uber.org/zap"
)

type magicTextResponse struct {
	Status      string
	Code        string
	MessageID   string `json:"Message-Id"`
	Description string
}

func (s *Service) sendSMS(ctx context.Context, apiURL string) error {
	l := ctx.Logger()
	l.Debug("sending MagicText SMS", zap.String("url", apiURL))
	resp, err := s.c.Get(apiURL)
	if err != nil {
		l.Error("failed to send MagicText SMS", zap.Error(err))
		return fmt.Errorf("failed to send MagicText SMS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		l.Error("failed to send MagicText SMS", zap.Int("status_code", resp.StatusCode), zap.String("response", string(body)))
		return fmt.Errorf("failed to send MagicText SMS, status code: %d", resp.StatusCode)
	}

	respData := magicTextResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		l.Error("failed to decode MagicText SMS response", zap.Error(err), zap.String("response", string(body)))
		return fmt.Errorf("failed to decode MagicText SMS response: %w", err)
	}

	s.l.Debug("MagicText SMS sent sucessfully", zap.String("code", respData.Code), zap.String("message_id", respData.MessageID))
	return nil
}

func (s *Service) MagicTextSMS(ctx context.Context, msg *Message) error {
	l := ctx.Logger()
	if len(msg.To) < 1 {
		l.Error("atleast one recipient phone number required", zap.Int("count", len(msg.To)))
		return fmt.Errorf("atleast one recipient phone number required")
	}

	_, err := govalidator.ValidateStruct(msg)
	if err != nil {
		l.Error("invalid message", zap.Error(err))
		return fmt.Errorf("invalid message: %w", err)
	}

	mobile := getCleanMobile(msg.To[0])
	if len(msg.To) > 1 {
		mobile = getCleanMobiles(msg.To)
	}
	body := url.QueryEscape(msg.Body)

	apiURL := "http://panel.magictext.in/http-tokenkeyapi.php" +
		"?authentic-key=" + s.token +
		"&senderid=" + s.senderID + "&route=1" +
		"&templateid=" + msg.TemplateID +
		"&number=" + mobile +
		"&message=" + body

	err = s.sendSMS(ctx, apiURL)
	if err != nil {
		l.Error("failed to send MagicText SMS", zap.Error(err))
		return err
	}

	return nil
}

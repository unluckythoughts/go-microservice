package text

import (
	"net/http"
	"strings"

	"github.com/unluckythoughts/go-microservice/v2/utils"
	"go.uber.org/zap"
)

type Service struct {
	c        *http.Client
	l        *zap.Logger
	token    string
	senderID string
	baseURL  string
}

type Options struct {
	// Token is the authentication token for the text messaging service.
	// It should be set in the environment variable TEXT_TOKEN.
	Token string `env:"TEXT_TOKEN" envDefault:""`
	// SenderID is the sender ID for the text messaging service.
	// It should be set in the environment variable TEXT_SENDER_ID.
	SenderID string `env:"TEXT_SENDER_ID" envDefault:""`
	// BaseURL is the base URL for the text messaging service API.
	// It should be set in the environment variable TEXT_BASE_URL.
	BaseURL string `env:"TEXT_BASE_URL" envDefault:""`
}

type Message struct {
	// To is the recipient's phone number. It must be a valid phone number in E.164 format (e.g., +1234567890).
	To []string `valid:"required~recipient phone numbers are required,indian_mobiles~invalid phone numbers"`
	// TemplateID is the ID of the message template to use when sending the text message.
	// It is required and cannot be empty. The template ID should correspond to a valid template in the text messaging service.
	TemplateID string `valid:"required~message template ID is required"`
	// Body is the content of the text message. It is required and cannot be empty.
	Body string `valid:"required~text message body is required"`
}

func defaultOptions(overrides *Options) *Options {
	opts := Options{}

	utils.ParseEnvironmentVars(&opts)

	if overrides != nil {
		if overrides.Token != "" {
			opts.Token = overrides.Token
		}
		if overrides.SenderID != "" {
			opts.SenderID = overrides.SenderID
		}
		if overrides.BaseURL != "" {
			opts.BaseURL = overrides.BaseURL
		}
	}

	return &opts
}

func New(overrides *Options) *Service {
	opts := defaultOptions(overrides)

	return &Service{
		c:        &http.Client{},
		token:    opts.Token,
		senderID: opts.SenderID,
		baseURL:  opts.BaseURL,
	}
}

func getCleanMobile(mobile string) string {
	// Remove all non-digit characters from the mobile number
	cleaned := ""
	for _, r := range mobile {
		if r >= '0' && r <= '9' {
			cleaned += string(r)
		}
	}
	return cleaned
}

func getCleanMobiles(mobiles []string) string {
	cleanedMobiles := make([]string, len(mobiles))
	for i, mobile := range mobiles {
		mobile = getCleanMobile(mobile)
		if len(mobile) == 10 {
			mobile = "91" + mobile
		}
		cleanedMobiles[i] = mobile
	}
	return strings.Join(cleanedMobiles, ",")
}

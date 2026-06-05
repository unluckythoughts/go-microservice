package phonepe

import (
	"fmt"

	"github.com/PhonePe/phonepe-pg-sdk-go/common/types"
	"github.com/PhonePe/phonepe-pg-sdk-go/payments/v2/standardcheckout"
	subscription "github.com/PhonePe/phonepe-pg-sdk-go/subscription/v2"
	"github.com/unluckythoughts/go-microservice/v2/utils"
)

type Options struct {
	// Environment is the environment to use for API calls. Can be "production" or "sandbox". Default is "sandbox"
	Environment string `env:"PHONEPE_ENVIRONMENT" envDefault:"sandbox"`
	// MerchantID is the merchant ID provided by PhonePe for authentication
	MerchantID string `env:"PHONEPE_MERCHANT_ID" envDefault:""`
	// ClientID is the client ID provided by PhonePe for authentication
	ClientID string `env:"PHONEPE_CLIENT_ID" envDefault:""`
	// ClientSecret is the client secret provided by PhonePe for authentication
	ClientSecret string `env:"PHONEPE_CLIENT_SECRET" envDefault:""`
	// BaseURL is the base URL for API calls. If not provided, it will be set based on the environment
	BaseURL string `env:"PHONEPE_BASE_URL" envDefault:""`
	// RedirectURI is the redirect URI for authentication. If not provided, it will be set to a default value
	RedirectURI string `env:"PHONEPE_REDIRECT_URI" envDefault:""`
	// CallbackURL is the callback URL for authentication. If not provided, it will be set to a default value
	CallbackURL string `env:"PHONEPE_CALLBACK_URL" envDefault:""`
}

type Service struct {
	c  *standardcheckout.StandardCheckoutClient
	sc *subscription.SubscriptionClient
}

func defaultOptions(overrides Options) Options {
	opts := Options{}
	utils.ParseEnvironmentVars(&opts)

	if overrides.Environment != "" {
		opts.Environment = overrides.Environment
	}
	if overrides.MerchantID != "" {
		opts.MerchantID = overrides.MerchantID
	}
	if overrides.ClientID != "" {
		opts.ClientID = overrides.ClientID
	}
	if overrides.ClientSecret != "" {
		opts.ClientSecret = overrides.ClientSecret
	}
	if overrides.BaseURL != "" {
		opts.BaseURL = overrides.BaseURL
	}
	if overrides.RedirectURI != "" {
		opts.RedirectURI = overrides.RedirectURI
	}
	if overrides.CallbackURL != "" {
		opts.CallbackURL = overrides.CallbackURL
	}

	return opts
}

func New(opts Options) *Service {
	opts = defaultOptions(opts)

	env := types.Sandbox
	if opts.Environment == "production" {
		env = types.Production
	} else if opts.Environment == "test" {
		env = types.Test
	} else if opts.Environment != "sandbox" {
		panic("Invalid environment. Must be one of 'production', 'sandbox', or 'test'")
	}

	client, err := standardcheckout.GetInstance(
		opts.ClientID,
		opts.ClientSecret,
		1,
		env,
		false,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create PhonePe client: %v", err))
	}

	subClient, err := subscription.GetInstance(
		opts.ClientID,
		opts.ClientSecret,
		1,
		env,
		false,
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create PhonePe subscription client: %v", err))
	}

	return &Service{
		c:  client,
		sc: subClient,
	}
}

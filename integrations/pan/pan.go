package pan

import (
	"net/http"

	"github.com/unluckythoughts/go-microservice/v2/utils"
)

type Service struct {
	c     *http.Client
	code  string
	token string
}

type PANDetails struct {
	Number string
	Name   string
	Status string
}

type Options struct {
	// Code is the code provided by service provider for authentication
	Code string `env:"PAN_CODE" envDefault:""`
	// Token is the token / api key provided by service provider for authentication
	Token string `env:"PAN_TOKEN" envDefault:""`
}

func defaultOptions(overrides Options) Options {
	opts := Options{}
	utils.ParseEnvironmentVars(&opts)

	if overrides.Code != "" {
		opts.Code = overrides.Code
	}
	if overrides.Token != "" {
		opts.Token = overrides.Token
	}
	return opts
}

func New(opts Options) *Service {
	opts = defaultOptions(opts)

	return &Service{
		c:     &http.Client{},
		code:  opts.Code,
		token: opts.Token,
	}
}

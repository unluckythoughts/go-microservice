package utils

import (
	"github.com/caarlos0/env"
	"github.com/pkg/errors"
)

func ParseEnvironmentVars(config interface{}) {
	if err := env.Parse(config); err != nil {
		panic(errors.Wrapf(err, "could not parse env variables"))
	}
}

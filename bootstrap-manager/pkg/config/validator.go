package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

// use a single instance of Validate, it caches struct info
var v *validator.Validate

func Validate(cfg *Config) error {

	logrus.Infoln("validating configuration...")

	v = validator.New()

	err := v.Struct(cfg)
	return err
}

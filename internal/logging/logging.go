package logging

import "go.uber.org/zap"

func New(environment string, level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	if environment == "development" {
		cfg = zap.NewDevelopmentConfig()
	}

	if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
		return nil, err
	}

	return cfg.Build()
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

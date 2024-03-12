package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Postgres struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Database string `yaml:"db"`
	} `yaml:"postgres"`
	NatsStreaming struct {
		Host    string `yaml:"host"`
		Port    int    `yaml:"port"`
		Cluster string `yaml:"cluster"`
		Client  string `yaml:"client"`
		Channel string `yaml:"channel"`
	} `yaml:"nats-streaming"`
	ZapConfig zap.Config
}

func NewConfig(filename string) (*AppConfig, error) {
	var config AppConfig

	configFile, err := os.ReadFile(filename)
	if err != nil {
		return &config, err
	}

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	config.ZapConfig = zapConfig

	err = yaml.Unmarshal(configFile, &config)
	return &config, err
}

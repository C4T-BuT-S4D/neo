package client

import (
	"fmt"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func UnmarshalConfig() (*Config, error) {
	cfg := new(Config)
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("parsing config structure: %w", err)
	}
	zap.L().Debug("Unmarshalled config", zap.Any("config", cfg))
	return cfg, nil
}

type Config struct {
	Host        string `mapstructure:"host"`
	ExploitDir  string `mapstructure:"exploit_dir"`
	GrpcAuthKey string `mapstructure:"grpc_auth_key"`
	UseTLS      bool   `mapstructure:"use_tls"`
}

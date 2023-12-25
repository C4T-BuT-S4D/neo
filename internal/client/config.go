package client

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func MustUnmarshalConfig() *Config {
	cfg := new(Config)
	if err := viper.Unmarshal(&cfg); err != nil {
		logrus.Fatalf("Could not parse config structure: %v", err)
	}
	logrus.Debugf("Unmarshalled config %+v", cfg)
	return cfg
}

type Metrics struct {
	URL      string `mapstructure:"url"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type Config struct {
	Host        string `mapstructure:"host"`
	ExploitDir  string `mapstructure:"exploit_dir"`
	GrpcAuthKey string `mapstructure:"grpc_auth_key"`
	UseTLS      bool   `mapstructure:"use_tls"`

	Metrics *Metrics `mapstructure:"metrics"`
}

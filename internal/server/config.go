package server

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Config struct {
	Debug            bool              `mapstructure:"debug"`
	Port             string            `mapstructure:"port"`
	DBPath           string            `mapstructure:"db_path"`
	RedisURL         string            `mapstructure:"redis_url"`
	BaseDir          string            `mapstructure:"base_dir"`
	PingEvery        time.Duration     `mapstructure:"ping_every"`
	SubmitEvery      time.Duration     `mapstructure:"submit_every"`
	FarmConfig       FarmConfig        `mapstructure:"farm"`
	GrpcAuthKey      string            `mapstructure:"grpc_auth_key"`
	Environ          map[string]string `mapstructure:"env"`
	MetricsNamespace string            `mapstructure:"metrics_namespace"`
}

type FarmConfig struct {
	URL        string            `mapstructure:"url"`
	Password   string            `mapstructure:"password"`
	FlagRegexp string            `json:"FLAG_FORMAT"`
	Teams      map[string]string `json:"TEAMS"`
}

func (cfg *FarmConfig) ParseJSON(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(cfg); err != nil {
		return fmt.Errorf("decoding json: %w", err)
	}
	return nil
}

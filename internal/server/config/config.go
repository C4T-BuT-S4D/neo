package config

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type Config struct {
	LogLevel        string            `mapstructure:"log_level" default:"info"`
	Address         string            `mapstructure:"address" default:":5005"`
	StaticDir       string            `mapstructure:"static_dir" default:"front/dist"`
	DBPath          string            `mapstructure:"db_path" default:"data/db.db"`
	VictoriaLogsURL string            `mapstructure:"victorialogs_url" default:"http://victoria-logs:9428"`
	BaseDir         string            `mapstructure:"base_dir" default:"data/exploits"`
	PingEvery       time.Duration     `mapstructure:"ping_every" default:"5s"`
	SubmitEvery     time.Duration     `mapstructure:"submit_every" default:"2s"`
	FarmConfig      FarmConfig        `mapstructure:"farm"`
	GrpcAuthKey     string            `mapstructure:"grpc_auth_key"`
	Environ         map[string]string `mapstructure:"env"`

	MetricsAddress     string `mapstructure:"metrics_address" default:":3000"`
	MetricsNamespace   string `mapstructure:"metrics_namespace"`
	VictoriaMetricsURL string `mapstructure:"victoriametrics_url" default:"http://victoria:8428"`

	ConfigFile string `mapstructure:"config" default:"server_config.yml"`
}

type FarmConfig struct {
	URL        string            `json:"url" mapstructure:"url"`
	Password   string            `json:"password" mapstructure:"password"`
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

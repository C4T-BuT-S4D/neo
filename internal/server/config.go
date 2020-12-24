package server

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

func ReadConfig(data []byte, cfg *Config) error {
	return yaml.Unmarshal(data, cfg)
}

type Config struct {
	Port        string            `yaml:"port"`
	DBPath      string            `yaml:"db_path"`
	BaseDir     string            `yaml:"base_dir"`
	PingEvery   time.Duration     `yaml:"ping_every"`
	RunEvery    time.Duration     `yaml:"run_every"`
	Timeout     time.Duration     `yaml:"timeout"`
	FarmConfig  FarmConfig        `yaml:"farm"`
	GrpcAuthKey string            `yaml:"grpc_auth_key"`
	Environ     map[string]string `yaml:"var.env"`
}

type FarmConfig struct {
	Url        string            `yaml:"url"`
	Password   string            `yaml:"password"`
	FlagRegexp string            `json:"FLAG_FORMAT"`
	Teams      map[string]string `json:"TEAMS"`
}

func (cfg *FarmConfig) ParseJson(r io.Reader) error {
	dec := json.NewDecoder(r)
	if err := dec.Decode(cfg); err != nil {
		return fmt.Errorf("decoding json: %w", err)
	}
	return nil
}

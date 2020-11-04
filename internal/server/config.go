package server

import (
	"time"

	"gopkg.in/yaml.v3"
)

func ReadConfig(data []byte, cfg *Configuration) error {
	return yaml.Unmarshal(data, cfg)
}

type Configuration struct {
	Port         string        `yaml:"port"`
	DBPath       string        `yaml:"db_path"`
	BaseDir      string        `yaml:"base_dir"`
	PingEvery    time.Duration `yaml:"ping_every"`
	IPList       []string      `yaml:"ip_list"`
	RunEvery     time.Duration `yaml:"run_every"`
	Timeout      time.Duration `yaml:"timeout"`
	FarmUrl      string        `yaml:"farm_url"`
	FarmPassword string        `yaml:"farm_password"`
	FlagRegexp   string        `yaml:"flag_regexp"`
	GrpcAuthKey  string        `yaml:"grpc_auth_key"`
}

package server

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Port       string        `yaml:"port"`
	DBPath     string        `yaml:"db_path"`
	BaseDir    string        `yaml:"base_dir"`
	PingEvery  time.Duration `yaml:"ping_every"`
	IPList     []string      `yaml:"ip_list"`
	RunEvery   time.Duration `yaml:"run_every"`
	Timeout    time.Duration `yaml:"timeout"`
	FarmUrl    string        `yaml:"farm_url"`
	FlagRegexp string        `yaml:"flag_regexp"`
}

func ReadConfig(data []byte, cfg *Configuration) error {
	return yaml.Unmarshal(data, cfg)
}

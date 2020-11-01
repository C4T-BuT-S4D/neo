package config

import (
	"github.com/sirupsen/logrus"
	"regexp"
	"time"

	neopb "neo/lib/genproto/neo"
)

type Config struct {
	PingEvery  time.Duration
	RunEvery   time.Duration
	Timeout    time.Duration
	FarmUrl    string
	FlagRegexp *regexp.Regexp
}

func ToProto(c *Config) *neopb.Config {
	return &neopb.Config{
		RunEvery:   c.RunEvery.String(),
		Timeout:    c.Timeout.String(),
		FarmUrl:    c.FarmUrl,
		FlagRegexp: c.FlagRegexp.String(),
		PingEvery:  c.PingEvery.String(),
	}
}

func FromProto(config *neopb.Config) (*Config, error) {
	var (
		cfg Config
		err error
	)
	if cfg.FlagRegexp, err = regexp.Compile(config.GetFlagRegexp()); err != nil {
		return nil, err
	}
	if cfg.PingEvery, err = time.ParseDuration(config.GetPingEvery()); err != nil {
		return nil, err
	}
	if cfg.RunEvery, err = time.ParseDuration(config.GetRunEvery()); err != nil {
		return nil, err
	}
	if cfg.Timeout, err = time.ParseDuration(config.GetTimeout()); err != nil {
		return nil, err
	}
	cfg.FarmUrl = config.GetFarmUrl()
	logrus.Infof("Config timeout: %v", cfg.Timeout)
	return &cfg, nil
}

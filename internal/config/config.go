package config

import (
	"fmt"
	"regexp"
	"time"

	neopb "neo/lib/genproto/neo"

	"google.golang.org/protobuf/types/known/durationpb"
)

type Config struct {
	PingEvery    time.Duration
	SubmitEvery  time.Duration
	FarmURL      string
	FarmPassword string
	FlagRegexp   *regexp.Regexp
	Environ      []string
}

func ToProto(c *Config) *neopb.Config {
	return &neopb.Config{
		FarmUrl:      c.FarmURL,
		FarmPassword: c.FarmPassword,
		FlagRegexp:   c.FlagRegexp.String(),
		PingEvery:    durationpb.New(c.PingEvery),
		SubmitEvery:  durationpb.New(c.SubmitEvery),
		Environ:      c.Environ,
	}
}

func FromProto(config *neopb.Config) (*Config, error) {
	var (
		cfg Config
		err error
	)
	if cfg.FlagRegexp, err = regexp.Compile(config.FlagRegexp); err != nil {
		return nil, fmt.Errorf("compiling regex: %w", err)
	}
	cfg.FarmURL = config.FarmUrl
	cfg.FarmPassword = config.FarmPassword
	cfg.PingEvery = config.PingEvery.AsDuration()
	cfg.SubmitEvery = config.SubmitEvery.AsDuration()
	cfg.Environ = config.Environ
	return &cfg, nil
}

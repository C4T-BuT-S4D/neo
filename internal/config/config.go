package config

import (
	"fmt"
	"regexp"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	epb "github.com/c4t-but-s4d/neo/v2/proto/go/exploits"
)

type ExploitsConfig struct {
	PingEvery    time.Duration
	SubmitEvery  time.Duration
	FarmURL      string
	FarmPassword string
	FlagRegexp   *regexp.Regexp
	Environ      []string
}

func ToProto(c *ExploitsConfig) *epb.Config {
	return &epb.Config{
		FarmUrl:      c.FarmURL,
		FarmPassword: c.FarmPassword,
		FlagRegexp:   c.FlagRegexp.String(),
		PingEvery:    durationpb.New(c.PingEvery),
		SubmitEvery:  durationpb.New(c.SubmitEvery),
		Environ:      c.Environ,
	}
}

func FromProto(config *epb.Config) (*ExploitsConfig, error) {
	var (
		cfg ExploitsConfig
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

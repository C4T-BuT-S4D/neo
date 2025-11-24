package config

import (
	"fmt"
	"regexp"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"

	epb "github.com/c4t-but-s4d/neo/v2/pkg/proto/exploits"
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
	if cfg.FlagRegexp, err = regexp.Compile(config.GetFlagRegexp()); err != nil {
		return nil, fmt.Errorf("compiling regex: %w", err)
	}
	cfg.FarmURL = config.GetFarmUrl()
	cfg.FarmPassword = config.GetFarmPassword()
	cfg.PingEvery = config.GetPingEvery().AsDuration()
	cfg.SubmitEvery = config.GetSubmitEvery().AsDuration()
	cfg.Environ = config.GetEnviron()
	return &cfg, nil
}

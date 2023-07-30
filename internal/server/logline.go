package server

import (
	"fmt"
	"strconv"

	"github.com/mitchellh/mapstructure"

	neopb "neo/lib/genproto/neo"
)

func NewLogLineFromValues(vals map[string]interface{}) (*LogLine, error) {
	var line LogLine
	if err := mapstructure.Decode(vals, &line); err != nil {
		return nil, fmt.Errorf("decoding structure: %w", err)
	}
	return &line, nil
}

func NewLogLineFromProto(p *neopb.LogLine) *LogLine {
	return &LogLine{
		Exploit: p.Exploit,
		Version: strconv.FormatInt(p.Version, 10),
		Message: p.Message,
		Level:   p.Level,
		Team:    p.Team,
	}
}

type LogLine struct {
	Exploit string `mapstructure:"exploit"`
	Version string `mapstructure:"version"`
	Message string `mapstructure:"message"`
	Level   string `mapstructure:"level"`
	Team    string `mapstructure:"team"`
}

func (l *LogLine) String() string {
	return fmt.Sprintf("Line(%s.v%s)", l.Exploit, l.Version)
}

func (l *LogLine) DumpValues() (map[string]interface{}, error) {
	res := make(map[string]interface{})
	if err := mapstructure.Decode(l, &res); err != nil {
		return nil, fmt.Errorf("encoding structure: %w", err)
	}
	return res, nil
}

func (l *LogLine) EstimateSize() int {
	const (
		estNum     = 5
		estDenom   = 4
		structSize = 8 * 5
	)
	sizeEst := structSize + len(l.Exploit) + len(l.Version) + len(l.Message) + len(l.Level) + len(l.Team)
	return sizeEst * estNum / estDenom
}

func (l *LogLine) ToProto() (*neopb.LogLine, error) {
	version, err := strconv.ParseInt(l.Version, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("converting version (%v): %w", l.Version, err)
	}
	return &neopb.LogLine{
		Exploit: l.Exploit,
		Version: version,
		Message: l.Message,
		Level:   l.Level,
		Team:    l.Team,
	}, nil
}

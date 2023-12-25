package logstor

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/timestamppb"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

func NewLineFromRedis(vals map[string]any) (*Line, error) {
	var line Line
	if err := mapstructure.Decode(vals, &line); err != nil {
		return nil, fmt.Errorf("decoding structure: %w", err)
	}
	return &line, nil
}

func NewLineFromProto(p *logspb.LogLine) *Line {
	return &Line{
		Timestamp: p.Timestamp.AsTime(),
		Exploit:   p.Exploit,
		Version:   p.Version,
		Message:   p.Message,
		Level:     p.Level,
		Team:      p.Team,
	}
}

type Line struct {
	Timestamp time.Time `mapstructure:"timestamp"`
	Exploit   string    `mapstructure:"exploit"`
	Version   int64     `mapstructure:"version"`
	Message   string    `mapstructure:"message"`
	Level     string    `mapstructure:"level"`
	Team      string    `mapstructure:"team"`
}

func (l *Line) String() string {
	return fmt.Sprintf("Line(%s.v%s)", l.Exploit, l.Version)
}

func (l *Line) EstimateSize() int {
	const (
		estNum     = 5
		estDenom   = 4
		structSize = 8*4 + 32 + 8
	)
	sizeEst := structSize + len(l.Exploit) + len(l.Message) + len(l.Level) + len(l.Team)
	return sizeEst * estNum / estDenom
}

func (l *Line) ToRedis() (map[string]any, error) {
	res := make(map[string]any)
	if err := mapstructure.Decode(l, &res); err != nil {
		return nil, fmt.Errorf("encoding structure: %w", err)
	}
	return res, nil
}

func (l *Line) ToProto() *logspb.LogLine {
	return &logspb.LogLine{
		Timestamp: timestamppb.New(l.Timestamp),
		Exploit:   l.Exploit,
		Version:   l.Version,
		Message:   l.Message,
		Level:     l.Level,
		Team:      l.Team,
	}
}

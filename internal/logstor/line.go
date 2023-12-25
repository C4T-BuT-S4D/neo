package logstor

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"google.golang.org/protobuf/types/known/timestamppb"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

func LineDecodeHook(from, to reflect.Type, data any) (any, error) {
	switch {
	case from.Kind() == reflect.String && to == reflect.TypeOf(time.Time{}):
		str, ok := data.(string)
		if !ok {
			return data, nil
		}

		res, err := time.Parse(time.RFC3339Nano, str)
		if err != nil {
			return nil, fmt.Errorf("parsing time: %w", err)
		}
		return res, nil

	case from.Kind() == reflect.String && to.Kind() == reflect.Int64:
		str, ok := data.(string)
		if !ok {
			return data, nil
		}

		res, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing int64: %w", err)
		}
		return res, nil

	default:
		return data, nil
	}
}

func NewLineFromRedis(vals map[string]any) (*Line, error) {
	var line Line
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: LineDecodeHook,
		Result:     &line,
	})
	if err != nil {
		return nil, fmt.Errorf("creating decoder: %w", err)
	}

	if err := decoder.Decode(vals); err != nil {
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

func (l *Line) ToRedis() map[string]any {
	return map[string]any{
		"timestamp": l.Timestamp.Format(time.RFC3339Nano),
		"exploit":   l.Exploit,
		"version":   strconv.FormatInt(l.Version, 10),
		"message":   l.Message,
		"level":     l.Level,
		"team":      l.Team,
	}
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

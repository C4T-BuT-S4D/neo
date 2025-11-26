package logstor

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/timestamppb"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

func TestParseVictoriaLogsLine(t *testing.T) {
	tests := []struct {
		name    string
		doc     map[string]any
		want    *logspb.LogLine
		wantErr bool
	}{
		{
			name: "valid line",
			doc: map[string]any{
				"_time":   "2024-01-01T12:00:00Z",
				"exploit": "test-exploit",
				"version": "1",
				"_msg":    "test message",
				"level":   "info",
				"team":    "team1",
			},
			want: &logspb.LogLine{
				Timestamp: timestamppb.New(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
				Exploit:   "test-exploit",
				Version:   1,
				Message:   "test message",
				Level:     "info",
				Team:      "team1",
			},
			wantErr: false,
		},
		{
			name: "missing _time field",
			doc: map[string]any{
				"exploit": "test-exploit",
				"version": "1",
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "invalid version",
			doc: map[string]any{
				"_time":   "2024-01-01T12:00:00Z",
				"exploit": "test-exploit",
				"version": "invalid",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVictoriaLogsLine(tt.doc)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if diff := cmp.Diff(tt.want, got, protocmp.Transform()); diff != "" {
				t.Errorf("parseVictoriaLogsLine() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBuildLogsQLQuery(t *testing.T) {
	tests := []struct {
		name    string
		exploit string
		version int64
		want    string
	}{
		{
			name:    "both exploit and version",
			exploit: "test-exploit",
			version: 1,
			want:    `exploit:"test-exploit" AND version:"1"`,
		},
		{
			name:    "only exploit",
			exploit: "test-exploit",
			version: 0,
			want:    `exploit:"test-exploit"`,
		},
		{
			name:    "empty",
			exploit: "",
			version: 0,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildLogsQLQuery(tt.exploit, tt.version)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestVictoriaLogsDocument(t *testing.T) {
	line := &logspb.LogLine{
		Timestamp: timestamppb.New(time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)),
		Exploit:   "test-exploit",
		Version:   1,
		Message:   "test message",
		Level:     "info",
		Team:      "team1",
	}

	doc := victoriaLogsDocument{
		Timestamp: line.GetTimestamp().AsTime().Format(time.RFC3339Nano),
		Exploit:   line.GetExploit(),
		Version:   "1",
		Message:   line.GetMessage(),
		Level:     line.GetLevel(),
		Team:      line.GetTeam(),
	}

	require.Equal(t, "2024-01-01T12:00:00.000000000Z", doc.Timestamp)
	require.Equal(t, "test-exploit", doc.Exploit)
	require.Equal(t, "1", doc.Version)
	require.Equal(t, "test message", doc.Message)
	require.Equal(t, "info", doc.Level)
	require.Equal(t, "team1", doc.Team)
}

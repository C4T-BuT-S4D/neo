package logstor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

var _ Storage = (*VictoriaLogsStorage)(nil)

type VictoriaLogsStorage struct {
	baseURL string
	client  *http.Client
}

type victoriaLogsDocument struct {
	Timestamp string `json:"_time"`
	Exploit   string `json:"exploit"`
	Version   string `json:"version"`
	Message   string `json:"_msg"`
	Level     string `json:"level"`
	Team      string `json:"team"`
}

func NewVictoriaLogsStorage(ctx context.Context, baseURL string) (*VictoriaLogsStorage, error) {
	// Ensure the URL is valid
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid base url: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &VictoriaLogsStorage{
		baseURL: baseURL,
		client:  client,
	}, nil
}

func (s *VictoriaLogsStorage) Add(ctx context.Context, lines ...*logspb.LogLine) error {
	if len(lines) == 0 {
		return nil
	}

	// Convert lines to NDJSON format
	var buf bytes.Buffer
	for _, line := range lines {
		doc := victoriaLogsDocument{
			Timestamp: line.GetTimestamp().AsTime().Format(time.RFC3339Nano),
			Exploit:   line.GetExploit(),
			Version:   strconv.FormatInt(line.GetVersion(), 10),
			Message:   line.GetMessage(),
			Level:     line.GetLevel(),
			Team:      line.GetTeam(),
		}
		if err := json.NewEncoder(&buf).Encode(doc); err != nil {
			return fmt.Errorf("encoding line %v: %w", line, err)
		}
	}

	// Send to VictoriaLogs
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/insert/jsonline", &buf)
	if err != nil {
		return fmt.Errorf("creating insert request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-ndjson")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending insert request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("insert failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (s *VictoriaLogsStorage) Search(ctx context.Context, searchReq *logspb.SearchLogLinesRequest) iter.Seq2[*logspb.LogLine, error] {
	return func(yield func(*logspb.LogLine, error) bool) {
		query := buildLogsQLQuery(searchReq.GetExploit(), searchReq.GetVersion())
		if query != "" {
			query += " | sort by (_time)"
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/select/logsql/query", nil)
		if err != nil {
			yield(nil, fmt.Errorf("creating search request: %w", err))
			return
		}

		q := req.URL.Query()
		q.Set("query", query)

		limit := searchReq.GetLimit()
		if limit == 0 {
			limit = 10000 // default limit
		}
		q.Set("limit", strconv.FormatInt(limit, 10))

		if searchReq.GetLastToken() != "" {
			// lastToken for VictoriaLogs is a timestamp
			q.Set("start", searchReq.GetLastToken())
		}
		req.URL.RawQuery = q.Encode()

		resp, err := s.client.Do(req)
		if err != nil {
			yield(nil, fmt.Errorf("sending search request: %w", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			yield(nil, fmt.Errorf("search failed with status %d: %s", resp.StatusCode, string(body)))
			return
		}

		// Parse response - VictoriaLogs returns NDJSON
		decoder := json.NewDecoder(resp.Body)

		for decoder.More() {
			var doc map[string]interface{}
			if err := decoder.Decode(&doc); err != nil {
				if err == io.EOF {
					break
				}
				yield(nil, fmt.Errorf("decoding line: %w", err))
				return
			}

			line, err := parseVictoriaLogsLine(doc)
			if err != nil {
				// Skip invalid lines
				continue
			}

			if !yield(line, nil) {
				return
			}
		}
	}
}

func parseVictoriaLogsLine(doc map[string]interface{}) (*logspb.LogLine, error) {
	getStr := func(key string) string {
		if v, ok := doc[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	timestampStr := getStr("_time")
	if timestampStr == "" {
		return nil, fmt.Errorf("missing _time field")
	}

	timestamp, err := time.Parse(time.RFC3339Nano, timestampStr)
	if err != nil {
		timestamp, err = time.Parse(time.RFC3339, timestampStr)
		if err != nil {
			return nil, fmt.Errorf("parsing timestamp: %w", err)
		}
	}

	versionStr := getStr("version")
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parsing version: %w", err)
	}

	return &logspb.LogLine{
		Timestamp: timestamppb.New(timestamp),
		Exploit:   getStr("exploit"),
		Version:   version,
		Message:   getStr("_msg"),
		Level:     getStr("level"),
		Team:      getStr("team"),
	}, nil
}

// buildLogsQLQuery builds a LogsQL query string
func buildLogsQLQuery(exploit string, version int64) string {
	var parts []string
	if exploit != "" {
		parts = append(parts, fmt.Sprintf(`exploit:"%s"`, exploit))
	}
	if version != 0 {
		parts = append(parts, fmt.Sprintf(`version:"%d"`, version))
	}
	return strings.Join(parts, " AND ")
}

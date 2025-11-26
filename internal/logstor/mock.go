package logstor

import (
	"context"
	"iter"
	"sync"

	logspb "github.com/c4t-but-s4d/neo/v2/pkg/proto/logs"
)

var _ Storage = (*MockStorage)(nil)

type MockStorage struct {
	mu    sync.RWMutex
	lines []*logspb.LogLine

	AddErr    error
	SearchErr error
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		lines: make([]*logspb.LogLine, 0),
	}
}

func (m *MockStorage) Add(_ context.Context, lines ...*logspb.LogLine) error {
	if m.AddErr != nil {
		return m.AddErr
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.lines = append(m.lines, lines...)
	return nil
}

func (m *MockStorage) Search(_ context.Context, req *logspb.SearchLogLinesRequest) iter.Seq2[*logspb.LogLine, error] {
	return func(yield func(*logspb.LogLine, error) bool) {
		if m.SearchErr != nil {
			yield(nil, m.SearchErr)
			return
		}

		m.mu.RLock()
		defer m.mu.RUnlock()

		count := int64(0)
		limit := req.GetLimit()
		if limit == 0 {
			limit = 10000
		}

		for _, line := range m.lines {
			if req.GetExploit() != "" && line.GetExploit() != req.GetExploit() {
				continue
			}
			if req.GetVersion() != 0 && line.GetVersion() != req.GetVersion() {
				continue
			}

			count++
			if count > limit {
				return
			}

			if !yield(line, nil) {
				return
			}
		}
	}
}

func (m *MockStorage) Lines() []*logspb.LogLine {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*logspb.LogLine, len(m.lines))
	copy(result, m.lines)
	return result
}

func (m *MockStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.lines = m.lines[:0]
}

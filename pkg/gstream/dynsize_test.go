package gstream

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicSizeCache(t *testing.T) {
	s := &mockWStream{}
	cache := NewDynamicSizeCache[*mockSizable, []*mockSizable](
		s,
		10,
		func(a []*mockSizable) (*[]*mockSizable, error) {
			return &a, nil
		},
	)
	gen := func(a int) *mockSizable {
		return &mockSizable{size: a}
	}
	require.NoError(t, cache.Queue(gen(5), gen(3)))
	require.Empty(t, s.sent)

	require.NoError(t, cache.Flush())
	require.Equal(t, [][]*mockSizable{{gen(5), gen(3)}}, s.sent)
	s.sent = nil

	require.NoError(t, cache.Queue(gen(5), gen(3)))
	require.NoError(t, cache.Queue(gen(2), gen(3)))
	require.Equal(t, [][]*mockSizable{{gen(5), gen(3), gen(2)}}, s.sent)
	require.NoError(t, cache.Flush())
	require.Equal(t, [][]*mockSizable{{gen(5), gen(3), gen(2)}, {gen(3)}}, s.sent)
}

type mockSizable struct {
	size int
}

func (s *mockSizable) EstimateSize() int {
	return s.size
}

type mockWStream struct {
	sent [][]*mockSizable
}

func (m *mockWStream) Send(t *[]*mockSizable) error {
	m.sent = append(m.sent, *t)
	return nil
}

func (m *mockWStream) Context() context.Context {
	return context.Background()
}

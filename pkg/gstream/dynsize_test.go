package gstream

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicSizeCacheSimple(t *testing.T) {
	s := &mockWStream{}
	cache := NewDynamicSizeCache[*mockSizable, []*mockSizable](
		s,
		10,  // maxSize
		100, // maxCount
		func(m *mockSizable) int { return m.size },
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

func TestDynamicSizeCache_ErrorPropagation(t *testing.T) {
	mockErr := errors.New("mock error")
	s := &mockWStream{returnErr: mockErr}
	cache := NewDynamicSizeCache[*mockSizable, []*mockSizable](
		s,
		5,   // maxSize
		100, // maxCount
		func(m *mockSizable) int { return m.size },
		func(a []*mockSizable) (*[]*mockSizable, error) {
			return &a, nil
		},
	)
	err := cache.Queue(&mockSizable{size: 5})
	require.ErrorIs(t, err, mockErr)
	require.NotEqual(t, mockErr, err)
	require.Equal(t, [][]*mockSizable{{&mockSizable{size: 5}}}, s.sent)
}

func TestDynamicSizeCache_MaxCount(t *testing.T) {
	s := &mockWStream{}
	cache := NewDynamicSizeCache[*mockSizable, []*mockSizable](
		s,
		1000, // maxSize - large enough to not trigger
		3,    // maxCount - flush after 3 items
		func(m *mockSizable) int { return m.size },
		func(a []*mockSizable) (*[]*mockSizable, error) {
			return &a, nil
		},
	)
	gen := func(a int) *mockSizable {
		return &mockSizable{size: a}
	}

	// Queue 2 items - should not flush
	require.NoError(t, cache.Queue(gen(1), gen(1)))
	require.Empty(t, s.sent)

	// Queue 1 more item - should flush (total 3)
	require.NoError(t, cache.Queue(gen(1)))
	require.Equal(t, [][]*mockSizable{{gen(1), gen(1), gen(1)}}, s.sent)

	// Queue 4 more items - should flush after 3
	s.sent = nil
	require.NoError(t, cache.Queue(gen(2), gen(2), gen(2), gen(2)))
	require.Equal(t, [][]*mockSizable{{gen(2), gen(2), gen(2)}}, s.sent)

	// Final flush should get the remaining item
	require.NoError(t, cache.Flush())
	require.Equal(t, [][]*mockSizable{{gen(2), gen(2), gen(2)}, {gen(2)}}, s.sent)
}

type mockSizable struct {
	size int
}

type mockWStream struct {
	sent      [][]*mockSizable
	returnErr error
}

func (m *mockWStream) Send(t *[]*mockSizable) error {
	tmp := make([]*mockSizable, len(*t))
	copy(tmp, *t)
	m.sent = append(m.sent, tmp)
	return m.returnErr
}

func (m *mockWStream) Context() context.Context {
	return context.Background()
}

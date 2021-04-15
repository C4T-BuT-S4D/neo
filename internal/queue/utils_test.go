package queue

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type errorReader struct{}

func (r *errorReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func Test_safeReadOutput(t *testing.T) {
	bufferReader := func(s string) io.Reader {
		return bytes.NewBufferString(s)
	}
	rep := func(s string, cnt int) string {
		return strings.Repeat(s, cnt)
	}

	tests := []struct {
		name       string
		r          io.Reader
		wantChunks []string
		wantErr    error
	}{
		{"simple", bufferReader("test"), []string{"test"}, nil},
		{"empty", bufferReader(""), nil, nil},
		{"error", &errorReader{}, nil, io.ErrUnexpectedEOF},
		{
			"large chunk",
			bufferReader(rep("a", chunkSize) + rep("b", chunkSize) + rep("c", chunkSize)),
			[]string{
				rep("a", chunkSize) + rep("b", smallChunkSize),
				rep("b", chunkSize) + rep("c", smallChunkSize),
				rep("c", chunkSize),
			},
			nil,
		},
	}
	for _, tt := range tests {
		var chunks []string
		cb := func(buf []byte) {
			chunks = append(chunks, string(buf))
		}
		t.Run(tt.name, func(t *testing.T) {
			if err := safeReadOutput(tt.r, cb); !errors.Is(err, tt.wantErr) {
				t.Errorf("safeReadOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.wantChunks, chunks, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("safeReadOutout() mismatch chunks (-want +got):\n%s", diff)
			}
		})
	}
}

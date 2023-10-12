package filestream

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	fspb "github.com/c4t-but-s4d/neo/v2/proto/go/fileserver"
)

type failedReadWriter struct {
}

var errTestWrite = errors.New("test read write error")

func (rw *failedReadWriter) Write(_ []byte) (n int, err error) {
	return 0, errTestWrite
}

func (rw *failedReadWriter) Read(_ []byte) (n int, err error) {
	return 0, errTestWrite
}

type mockUploadStream struct {
	withError bool
	buf       bytes.Buffer
}

func (ms *mockUploadStream) Send(s *fspb.FileStream) error {
	ms.buf.Write(s.Chunk)
	if ms.withError {
		return errTestWrite
	}
	return nil
}

type mockDownloadStream struct {
	withError bool
	data      *strings.Reader
	chunkSize int
}

func (ms *mockDownloadStream) Recv() (*fspb.FileStream, error) {
	if ms.chunkSize == 0 {
		ms.chunkSize = chunkSize
	}
	if ms.withError {
		return nil, errTestWrite
	}
	b := make([]byte, ms.chunkSize)
	n, err := ms.data.Read(b)
	if err != nil {
		return nil, fmt.Errorf("reading stream content: %w", err)
	}
	return &fspb.FileStream{Chunk: b[:n]}, nil
}

func TestLoad(t *testing.T) {
	for _, tc := range []struct {
		reader io.Reader
		stream *mockUploadStream
		want   string
		err    error
	}{
		{
			reader: strings.NewReader("somedata"),
			stream: &mockUploadStream{withError: false},
			want:   "somedata",
			err:    nil,
		},
		{
			reader: &failedReadWriter{},
			stream: &mockUploadStream{withError: false},
			want:   "",
			err:    errTestWrite,
		},
		{
			reader: strings.NewReader("somedata"),
			stream: &mockUploadStream{withError: true},
			want:   "somedata",
			err:    errTestWrite,
		},
	} {
		require.ErrorIs(t, Load(tc.reader, tc.stream), tc.err)
		if diff := cmp.Diff(tc.want, tc.stream.buf.String()); diff != "" {
			t.Errorf("Load() result mismatch (-want +got):\n%s", diff)
		}
	}
}

func TestSave(t *testing.T) {
	for _, tc := range []struct {
		writer *strings.Builder
		stream *mockDownloadStream
		want   string
		err    error
	}{
		{
			writer: &strings.Builder{},
			stream: &mockDownloadStream{data: strings.NewReader("abacaba")},
			want:   "abacaba",
			err:    nil,
		},
		{
			writer: &strings.Builder{},
			stream: &mockDownloadStream{data: strings.NewReader("abacaba"), chunkSize: 1},
			want:   "abacaba",
			err:    nil,
		},
		{
			writer: &strings.Builder{},
			stream: &mockDownloadStream{data: strings.NewReader("abacaba"), withError: true},
			want:   "",
			err:    errTestWrite,
		},
	} {
		require.ErrorIs(t, Save(tc.stream, tc.writer), tc.err)
		if diff := cmp.Diff(tc.want, tc.writer.String()); diff != "" {
			t.Errorf("Save() result mismatch (-want +got):\n%s", diff)
		}
	}
}

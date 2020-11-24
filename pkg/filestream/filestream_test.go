package filestream

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	neopb "neo/lib/genproto/neo"
)

type failedReadWriter struct {
}

var readWriteError = errors.New("test read write error")

func (rw *failedReadWriter) Write(_ []byte) (n int, err error) {
	return 0, readWriteError
}

func (rw *failedReadWriter) Read(_ []byte) (n int, err error) {
	return 0, readWriteError
}

type mockUploadStream struct {
	withError bool
	buf       bytes.Buffer
}

func (ms *mockUploadStream) Send(s *neopb.FileStream) error {
	ms.buf.Write(s.GetChunk())
	if ms.withError {
		return readWriteError
	}
	return nil
}

type mockDownloadStream struct {
	withError bool
	data      *strings.Reader
	chunkSize int
}

func (ms *mockDownloadStream) Recv() (*neopb.FileStream, error) {
	if ms.chunkSize == 0 {
		ms.chunkSize = chunkSize
	}
	if ms.withError {
		return &neopb.FileStream{}, readWriteError
	}
	b := make([]byte, ms.chunkSize)
	n, err := ms.data.Read(b)
	return &neopb.FileStream{Chunk: b[:n]}, err
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
			err:    readWriteError,
		},
		{
			reader: strings.NewReader("somedata"),
			stream: &mockUploadStream{withError: true},
			want:   "somedata",
			err:    readWriteError,
		},
	} {
		err := Load(tc.reader, tc.stream)
		if tc.err != err {
			t.Errorf("Load(): got unexpected error = %v", err)
		}
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
			err:    readWriteError,
		},
	} {
		err := Save(tc.stream, tc.writer)
		if tc.err != err {
			t.Errorf("Save(): got unexpected error = %v", err)
		}
		if diff := cmp.Diff(tc.want, tc.writer.String()); diff != "" {
			t.Errorf("Save() result mismatch (-want +got):\n%s", diff)
		}
	}
}

package filestream

import (
	"io"

	neopb "neo/lib/genproto/neo"
)

const (
	chunkSize = 4 * 1024 * 1024 // 2MB
)

type DownloadStream interface {
	Recv() (*neopb.FileStream, error)
}

type UploadStream interface {
	Send(*neopb.FileStream) error
}

func Save(stream DownloadStream, out io.Writer) error {
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := out.Write(in.GetChunk()); err != nil {
			return err
		}
	}
}

func Load(in io.Reader, stream UploadStream) error {
	b := make([]byte, chunkSize)
	for {
		n, err := in.Read(b)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(&neopb.FileStream{Chunk: b[:n]}); err != nil {
			return err
		}
	}
}

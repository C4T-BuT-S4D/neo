package fs

import (
	"fmt"
	"io"
	"os"
	"path"
)

type fileInterface interface {
	io.ReadWriteCloser
	Name() string
}

type filesystem interface {
	Create(string) (fileInterface, error)
	Open(string) (fileInterface, error)
}

type osFs struct {
	baseDir string
}

func newOsFs(dir string) (*osFs, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating dir %s: %w", dir, err)
	}
	return &osFs{baseDir: dir}, nil
}

func (o osFs) Create(f string) (fileInterface, error) {
	fi, err := os.Create(path.Join(o.baseDir, f))
	if err != nil {
		return nil, fmt.Errorf("creating file %s in %s: %w", f, o.baseDir, err)
	}
	return fi, nil
}

func (o osFs) Open(f string) (fileInterface, error) {
	fi, err := os.Open(path.Join(o.baseDir, f))
	if err != nil {
		return nil, fmt.Errorf("opening file %s in %s: %w", f, o.baseDir, err)
	}
	return fi, nil
}

package fs

import (
	"fmt"
	"io"
	"os"
	"path"

	"io/fs"
)

type filesystem interface {
	fs.FS

	Create(string) (NamedFile, error)
}

type NamedFile interface {
	io.WriteCloser
	Name() string
}

var _ filesystem = (*osFs)(nil)

type osFs struct {
	baseDir string
}

func newOsFs(dir string) (*osFs, error) {
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating dir %s: %w", dir, err)
	}
	return &osFs{baseDir: dir}, nil
}

func (o osFs) Create(f string) (NamedFile, error) {
	fi, err := os.Create(path.Join(o.baseDir, f))
	if err != nil {
		return nil, fmt.Errorf("creating file %s in %s: %w", f, o.baseDir, err)
	}
	return fi, nil
}

func (o osFs) Open(f string) (fs.File, error) {
	fi, err := os.Open(path.Join(o.baseDir, f))
	if err != nil {
		return nil, fmt.Errorf("opening file %s in %s: %w", f, o.baseDir, err)
	}
	return fi, nil
}

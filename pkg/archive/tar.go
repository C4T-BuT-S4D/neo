package archive

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"go.uber.org/zap"
)

func Untar(dst string, r io.Reader) error {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.Mkdir(dst, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dst, err)
		}
	}
	zr, err := zstd.NewReader(r)
	if err != nil {
		return fmt.Errorf("creating zstd reader: %w", err)
	}
	defer zr.Close()

	tr := tar.NewReader(zr)

	for {
		header, err := tr.Next()

		switch {
		case errors.Is(err, io.EOF):
			return nil

		case err != nil:
			return fmt.Errorf("parsing tar header: %w", err)

		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}

		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("opening file %s: %w", target, err)
			}

			if _, err := io.Copy(f, tr); err != nil {
				_ = f.Close()
				return fmt.Errorf("copying file content: %w", err)
			}

			if err := f.Close(); err != nil {
				return fmt.Errorf("closing file %s: %w", target, err)
			}
		}
	}
}

// Tar takes a source and variable writers and walks 'source' writing each file
// found to the tar writer.
func Tar(src string, w io.Writer) error {
	// ensure the src actually exists before trying to tar it
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("unable to tar files: %w", err)
	}

	zw, err := zstd.NewWriter(w)
	if err != nil {
		return fmt.Errorf("creating zstd writer: %w", err)
	}
	defer func() {
		if err := zw.Close(); err != nil {
			zap.L().Error("Error closing zstd writer", zap.Error(err))
		}
	}()

	tw := tar.NewWriter(zw)
	defer func() {
		if err := tw.Close(); err != nil {
			zap.L().Error("Error closing tar writer", zap.Error(err))
		}
	}()

	if err := filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return fmt.Errorf("creating tar file header: %w", err)
		}

		header.Name = strings.TrimPrefix(strings.ReplaceAll(file, src, ""), string(filepath.Separator))

		if err := tw.WriteHeader(header); err != nil {
			return fmt.Errorf("writing tar heeader: %w", err)
		}

		if fi.IsDir() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return fmt.Errorf("opening file %s: %w", file, err)
		}

		if _, err := io.Copy(tw, f); err != nil {
			_ = f.Close()
			return fmt.Errorf("copying file content: %w", err)
		}

		if err := f.Close(); err != nil {
			return fmt.Errorf("closing file: %w", err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("walking directory %s: %w", src, err)
	}
	return nil
}

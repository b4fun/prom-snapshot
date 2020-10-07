package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// CreateZipArchive creates a zip archive for snapshot.
type CreateZipArchive struct {
	Logger      logrus.FieldLogger
	SnapshotDir string
}

// CreateTo creates a zip archive stream.
func (c *CreateZipArchive) CreateTo(dest io.Writer) error {
	zipWriter := zip.NewWriter(dest)

	walkErr := filepath.Walk(c.SnapshotDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relativePath := strings.TrimPrefix(path, c.SnapshotDir)
		zipFile, err := zipWriter.Create(relativePath)
		if err != nil {
			return fmt.Errorf("create zip file %s: %w", path, err)
		}

		f, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}
		defer f.Close()

		_, err = io.Copy(zipFile, f)
		if err != nil {
			return fmt.Errorf("compress file %s: %w", path, err)
		}

		return nil
	})
	if walkErr != nil {
		return fmt.Errorf("compress snapshot %s failed: %w", c.SnapshotDir, walkErr)
	}
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("compress snapshot %s failed: %w", c.SnapshotDir, err)
	}

	return nil
}

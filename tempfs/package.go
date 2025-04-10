package tempfs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/leafbridge/leafbridge-deploy/filetime"
	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// Options hold a set of options for extraction directories.
type Options struct {
	// DeleteOnClose requests that temporary directories and their contents
	// are deleted when the directory is closed.
	DeleteOnClose bool
}

// ExtractionDir is an extraction directory for a package in LeafBridge.
//
// It is a temporary directory created via os.MkdirTemp. Its name will have
// "leafbridge-" as a prefix.
type ExtractionDir struct {
	path string
	dir  *os.Root
	opts Options
}

// OpenExtractionDirForPackage opens a temporary directory to receive
// extracted files from a package.
//
// It is the caller's responsibility to close the returned directory when
// finished with it.
//
// The options can be used to request that the returned directory is deleted
// when closed.
//
// TODO: Make the options variadic.
func OpenExtractionDirForPackage(pkg lbdeploy.PackageContent, opts Options) (ExtractionDir, error) {
	// Unfortunately, this returns a path instead of an open directory handle.
	dirPath, err := os.MkdirTemp("", "leafbridge-"+pkg.String())
	if err != nil {
		return ExtractionDir{}, err
	}

	// Sanity check the directory path to make sure it conforms to our
	// expectations. If it doesn't, then return an error.
	//
	// Note that We might call os.RemoveAll() on the path later, and we really
	// don't want to make that call on an unintended path, especially when
	// operating with SYSTEM privileges.
	if !strings.Contains(dirPath, "leafbridge") || !strings.Contains(dirPath, "Temp") {
		return ExtractionDir{}, fmt.Errorf("the os.MkdirTemp call failed to create a direcoty with the expected format: %s", dirPath)
	}

	// Open the root of the newly created temp directory.
	dir, err := os.OpenRoot(dirPath)
	if err != nil {
		return ExtractionDir{}, err
	}

	// Return the extraction directory.
	return ExtractionDir{
		path: dirPath,
		dir:  dir,
		opts: opts,
	}, nil
}

// Path returns the path to the extraction directory at the time of its
// creation.
func (d ExtractionDir) Path() string {
	return d.path
}

// MkdirAll ensures that the given relative directory path and all of its
// parents have been created within the extraction directory.
//
// If name does not identify a local file path, or if directory creation
// fails, it rturns an error.
func (d ExtractionDir) MkdirAll(path string) error {
	localized, err := filepath.Localize(path)
	if err != nil {
		return err
	}

	dirPath := filepath.Join(d.path, localized)

	os.MkdirAll(dirPath, 0755)

	return nil
}

// WriteFile reads data from r and writes it to the provided relative file
// path. It continues until the reader returns io.EOF or an error is
// encountered.
//
// If a non-zero modified time is provided, it is set as the file's
// modification time.
//
// The standard unix file separator, forward slash (/), must be used as the
// separator in the provided path.
func (d ExtractionDir) WriteFile(path string, r io.Reader, modified time.Time) (written int64, err error) {
	// Convert the unix-style path to windows-style and ensure that it is a
	// local path.
	localized, err := filepath.Localize(path)
	if err != nil {
		return 0, err
	}

	// If this file is in a subdirectory, open its parent.
	dirPath, fileName := filepath.Split(localized)
	var parent *os.Root
	if !strings.HasPrefix(dirPath, ".") {
		parent, err = d.dir.OpenRoot(dirPath)
		if err != nil {
			return 0, fmt.Errorf("failed to open parent directory: %w", err)
		}
		defer parent.Close()
	} else {
		parent = d.dir
	}

	// Create the file.
	file, err := parent.Create(fileName)
	if err != nil {
		return 0, fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write the file content.
	written, err = io.Copy(file, r)
	if err != nil {
		return written, err
	}

	// Preserve modification and creation dates, if available.
	if !modified.IsZero() {
		if err := filetime.SetFileModificationTime(file, modified); err != nil {
			return written, fmt.Errorf("failed to set modification time: %w", err)
		}
	}

	return written, err
}

// Close releases any file system resources consumed by the directory.
//
// If the directory was created with the DeleteOnClose option, calling this
// function will cause the directory and all of its contents to be deleted.
func (d ExtractionDir) Close() error {
	// Simple closure.
	if !d.opts.DeleteOnClose {
		return d.dir.Close()
	}

	// Close and delete.
	err1 := d.dir.Close()
	err2 := os.RemoveAll(d.path)

	// TODO: Use d.dir.RemoveAll() when Go 1.25 is released, which should
	// include it.

	return errors.Join(err1, err2)
}

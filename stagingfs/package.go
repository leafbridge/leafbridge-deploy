package stagingfs

import (
	"os"
	"path/filepath"

	"github.com/leafbridge/leafbridge-deploy/lbdeploy"
)

// PackageDir is a staging directory for a package in LeafBridge.
type PackageDir struct {
	content lbdeploy.PackageContent
	path    string
	dir     *os.Root
}

// OpenFile opens the staging file for the given package.
//
// It is the caller's responsibility to close the file when finished with it.
func (r PackageDir) OpenFile(pkg lbdeploy.Package) (PackageFile, error) {
	f, err := r.dir.OpenFile(pkg.FileName(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return PackageFile{}, err
	}
	return PackageFile{
		Name:   pkg.FileName(),
		Type:   pkg.Type,
		Format: pkg.Format,
		Path:   filepath.Join(r.path, pkg.FileName()),
		File:   f,
	}, nil
}

// Close releases any file handles or resources held by the package
// staging directory.
func (r PackageDir) Close() error {
	return r.dir.Close()
}

// PackageFile is an open file for a package.
type PackageFile struct {
	Name   string
	Type   lbdeploy.PackageType
	Format lbdeploy.PackageFormat
	Path   string
	*os.File
}

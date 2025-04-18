package lbdeploy

import (
	"fmt"
	"path/filepath"
	"slices"

	"golang.org/x/sys/windows"
)

// FileSystemResources describes resources accessed through the file system,
// either local or remote.
type FileSystemResources struct {
	Files       FileResourceMap      `json:"files,omitempty"`
	Directories DirectoryResourceMap `json:"directories,omitempty"`
}

// ResolveFile resolves the requested file resource, returning a file
// reference that can be mapped to a path on the local system.
//
// Successfully resolving a path means that its path on the local system
// can be determined, but it does not imply that the file exists.
//
// If the file cannot be resolved, an error is returned.
func (fs FileSystemResources) ResolveFile(file FileResourceID) (ref FileRef, err error) {
	// TODO: Consider making custom error types for resolution.

	// Look up the file by its ID.
	data, exists := fs.Files[file]
	if !exists {
		return FileRef{}, fmt.Errorf("the file id \"%s\" is not a declared resource in the deployment's file system resources", file)
	}

	// Successful resolution must end in a known folder.
	var root KnownFolder

	// Keep track of the directories we traverse, which will ultimately form
	// a lineage under the root.
	var lineage []DirectoryResource

	// Maintain a list of directory IDs that we've encountered, so that we
	// can detect cycles.
	seen := make(DirectoryResourceSet)

	// Start with the file's location and traverse its ancestry, recording
	// each parent along the way. Stop when we encounter a known folder.
	next := data.Location
	for {
		// Check for cycles.
		if seen.Contains(next) {
			return FileRef{}, fmt.Errorf("failed to resolve file \"%s\": the directory id \"%s\" has a cyclic reference to itself in the deployment's file system resources", file, next)
		}
		seen.Add(next)

		// Look for a directory with the ID.
		if dir, found := fs.Directories[next]; found {
			lineage = append(lineage, dir)
			if dir.Location == "" {
				return FileRef{}, fmt.Errorf("failed to resolve file \"%s\": the directory id \"%s\" does not have a location", file, next)
			}
			next = dir.Location
		}

		// Look for a known folder with the ID.
		if kf, found := GetKnownFolder(next); found {
			root = kf
			break
		}

		// The location is not defined.
		return FileRef{}, fmt.Errorf("failed to resolve file \"%s\": the directory id \"%s\" is not a declared resource in the deployment's file system resources", file, next)
	}

	// Reverse the order of the directories that were recorded, so that it
	// can easily be followed from top to bottom.
	slices.Reverse(lineage)

	return FileRef{
		Root:     root,
		Lineage:  lineage,
		FileID:   file,
		FilePath: data.Path,
	}, nil
}

// ResolveFile resolves the requested file resource, returning a file
// reference that can be mapped to a path on the local system.
//
// Successfully resolving a path means that its path on the local system
// can be determined, but it does not imply that the file exists.
//
// If the file cannot be resolved, an error is returned.
func (fs FileSystemResources) ResolveDirectory(dir DirectoryResourceID) (ref DirRef, err error) {
	// TODO: Consider making custom error types for resolution.

	// Look up the directory by its ID.
	data, exists := fs.Directories[dir]
	if !exists {
		return DirRef{}, fmt.Errorf("the directory id \"%s\" is not a declared resource in the deployment's file system resources", dir)
	}

	// Successful resolution must end in a known folder.
	var root KnownFolder

	// Keep track of the directories we traverse, which will ultimately form
	// a lineage under the root.
	var lineage []DirectoryResource

	// Maintain a list of directory IDs that we've encountered, so that we
	// can detect cycles.
	seen := make(DirectoryResourceSet)

	// Start with the directory's location and traverse its ancestry, recording
	// each parent along the way. Stop when we encounter a known folder.
	lineage = append(lineage, data)
	next := data.Location
	for {
		// Check for cycles.
		if seen.Contains(next) {
			return DirRef{}, fmt.Errorf("failed to resolve directory \"%s\": the directory id \"%s\" has a cyclic reference to itself in the deployment's file system resources", dir, next)
		}
		seen.Add(next)

		// Look for a directory with the ID.
		if dir, found := fs.Directories[next]; found {
			lineage = append(lineage, dir)
			if dir.Location == "" {
				return DirRef{}, fmt.Errorf("failed to resolve directory \"%s\": the directory id \"%s\" does not have a location", dir, next)
			}
			next = dir.Location
		}

		// Look for a known folder with the ID.
		if kf, found := GetKnownFolder(next); found {
			root = kf
			break
		}

		// The location is not defined.
		return DirRef{}, fmt.Errorf("failed to resolve directory \"%s\": the directory id \"%s\" is not a declared resource in the deployment's file system resources", dir, next)
	}

	// Reverse the order of the directories that were recorded, so that it
	// can easily be followed from top to bottom.
	slices.Reverse(lineage)

	return DirRef{
		Root:    root,
		Lineage: lineage,
	}, nil
}

// FileResourceMap holds a set of file resources mapped by their identifiers.
type FileResourceMap map[FileResourceID]FileResource

// FileResourceID is a unique identifier for a file resource.
type FileResourceID string

// FileResource describes a file resource.
type FileResource struct {
	Location DirectoryResourceID // A well-known directory, or another directory ID.
	Path     string              // Relative to location
}

// FileRef is a resolved reference to a file on the local file system.
type FileRef struct {
	Root     KnownFolder
	Lineage  []DirectoryResource
	FileID   FileResourceID
	FilePath string
}

// Dir returns a reference to the file's directory.
func (ref FileRef) Dir() DirRef {
	return DirRef{
		Root:    ref.Root,
		Lineage: ref.Lineage,
	}
}

// Path returns the path of the directory on the local file system.
func (ref FileRef) Path() (string, error) {
	path, err := ref.Dir().Path()
	if err != nil {
		return "", err
	}

	localized, err := filepath.Localize(ref.FilePath)
	if err != nil {
		return "", err
	}

	return filepath.Join(path, localized), nil
}

// DirectoryResourceMap holds a set of directory resources mapped by their
// identifiers.
type DirectoryResourceMap map[DirectoryResourceID]DirectoryResource

// DirectoryResourceID is a unique identifier for a directory resource.
type DirectoryResourceID string

// DirectoryType declares the type of a directory resource.
type DirectoryType string

// FileResource describes a directory resource.
type DirectoryResource struct {
	Location DirectoryResourceID // A well-known directory, or another directory ID.
	Path     string              // Relative to location
}

// DirRef is a resolved reference to a directory on the local file system.
type DirRef struct {
	Root    KnownFolder
	Lineage []DirectoryResource
}

// Path returns the path of the directory on the local file system.
func (ref DirRef) Path() (string, error) {
	root, err := ref.Root.Path()
	if err != nil {
		return "", err
	}

	path := root
	for _, dir := range ref.Lineage {
		localized, err := filepath.Localize(dir.Path)
		if err != nil {
			return "", err
		}
		path = filepath.Join(path, localized)
	}

	return path, nil
}

// DirectoryResourceSet holds a set of directory resource IDs.
type DirectoryResourceSet map[DirectoryResourceID]struct{}

// Contains returns true if the given id is present in the set.
func (set DirectoryResourceSet) Contains(id DirectoryResourceID) bool {
	_, present := set[id]
	return present
}

// Contains adds the give id to the set. If it is already present, it takes
// no action.
func (set DirectoryResourceSet) Add(id DirectoryResourceID) {
	set[id] = struct{}{}
}

// Remove removes the give id from the set. If it is not present, it takes
// no action.
func (set DirectoryResourceSet) Remove(id DirectoryResourceID) {
	delete(set, id)
}

// KnownFolderMap is a map of predefined directory resource IDs to known
// folder locations.
type KnownFolderMap map[DirectoryResourceID]KnownFolder

// KnownFolder is a folder with a known location.
type KnownFolder struct {
	id   DirectoryResourceID
	guid *windows.KNOWNFOLDERID
}

// ID returns the LeafBridge directory ID of the known folder.
func (kf KnownFolder) ID() DirectoryResourceID {
	return kf.id
}

// GUID returns the Known Folder ID in Windows.
func (kf KnownFolder) GUID() *windows.KNOWNFOLDERID {
	return kf.guid
}

// IsZero returns true if the known folder is undefined.
func (kf KnownFolder) IsZero() bool {
	return kf.guid == nil
}

// Path retrieves the path to the known folder on the local system.
func (kf KnownFolder) Path() (path string, err error) {
	path, err = windows.KnownFolderPath(kf.guid, 0)
	return
}

// GetKnownFolder looks for a known folder with the given directory resource
// ID. If one is found, it is returned and okay will be true.
func GetKnownFolder(id DirectoryResourceID) (folder KnownFolder, ok bool) {
	folder, ok = knownFolders[id]
	return
}

var knownFolders = KnownFolderMap{
	"program-data":   KnownFolder{guid: windows.FOLDERID_ProgramData, id: "program-data"},
	"start-menu":     KnownFolder{guid: windows.FOLDERID_CommonStartMenu, id: "start-menu"},
	"public-desktop": KnownFolder{guid: windows.FOLDERID_PublicDesktop, id: "public-desktop"},
}

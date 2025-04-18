package main

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/gentlemanautomaton/winobj/winmutex"
	"github.com/leafbridge/leafbridge-deploy/localfs"
)

// ShowCmd shows information that is relevant to a LeafBridge deployment.
type ShowCmd struct {
	Config    ShowConfigCmd    `kong:"cmd,help='Shows configuration loaded from a deployment configuration file.'"`
	Resources ShowResourcesCmd `kong:"cmd,help='Shows the current condition of relevant resources for a deployment.'"`
}

// ShowConfigCmd shows the configuration of a LeafBridge deployment.
type ShowConfigCmd struct {
	ConfigFile string `kong:"required,name='config-file',help='Path to a deployment file describing the deployment.'"`
}

// Run executes the LeafBridge show config command.
func (cmd ShowConfigCmd) Run(ctx context.Context) error {
	// Read the deployment file.
	dep, err := loadDeployment(cmd.ConfigFile)
	if err != nil {
		return err
	}

	// Print the loaded configuration.
	out, err := json.MarshalIndent(dep, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(out))

	return nil
}

// ShowResourcesCmd shows the current condition of relevant resources for
// a LeafBridge deployment.
type ShowResourcesCmd struct {
	ConfigFile string `kong:"required,name='config-file',help='Path to a deployment file describing the deployment.'"`
}

// Run executes the LeafBridge show resources command.
func (cmd ShowResourcesCmd) Run(ctx context.Context) error {
	// Read the deployment file.
	dep, err := loadDeployment(cmd.ConfigFile)
	if err != nil {
		return err
	}

	fmt.Printf("---- %s (%s) Resources ----\n", dep.Name, cmd.ConfigFile)

	// Mutex Resources.
	if mutexes := dep.Resources.Mutexes; len(mutexes) > 0 {
		// Sort the mutex IDs for a deterministic order.
		ids := slices.Collect(maps.Keys(mutexes))
		slices.Sort(ids)

		// Print information about each mutex.
		fmt.Printf("  Mutexes:\n")
		for _, id := range ids {
			mutex := mutexes[id]
			func() {
				fmt.Printf("    %s\n", id)

				// Print the object name of the mutex.
				name, err := mutex.ObjectName()
				if err != nil {
					fmt.Printf("      Name:     (%v)\n", err)
					return
				}
				fmt.Printf("      Name:     %s\n", name)

				// Print the status of the mutex.
				exists, err := winmutex.Exists(name)
				if err != nil {
					fmt.Printf("      Status:   (%v)\n", err)
					return
				}

				if exists {
					fmt.Printf("      Status:   Present\n")
				} else {
					fmt.Printf("      Status:   Missing\n")
				}
			}()
		}
	}

	// Directory Resources.
	if dirs := dep.Resources.FileSystem.Directories; len(dirs) > 0 {
		// Sort the directory IDs for a deterministic order.
		ids := slices.Collect(maps.Keys(dirs))
		slices.Sort(ids)

		// Print information about each file.
		fmt.Printf("  Directories:\n")
		for _, id := range ids {
			func() {
				fmt.Printf("    %s:\n", id)

				// Resolve the directory reference.
				ref, err := dep.Resources.FileSystem.ResolveDirectory(id)
				if err != nil {
					fmt.Printf("      Path:     %v\n", err)
					return
				}

				// Generate a file path.
				path, err := ref.Path()
				if err != nil {
					fmt.Printf("      Path:     %v\n", err)
					return
				}

				// Open the parent directory.
				dir, err := localfs.OpenDir(ref)
				if err != nil {
					fmt.Printf("      Path:     %s\n", path)
					if os.IsNotExist(err) {
						fmt.Printf("      Status:   Missing\n")
					} else {
						fmt.Printf("      Status:   %v\n", err)
					}
					return
				}
				defer dir.Close()

				// Print the path and status.
				fmt.Printf("      Path:     %s\n", dir.Path())
				fmt.Printf("      Status:   Present\n")
			}()
		}
	}

	// File Resources.
	if files := dep.Resources.FileSystem.Files; len(files) > 0 {
		// Sort the file IDs for a deterministic order.
		ids := slices.Collect(maps.Keys(files))
		slices.Sort(ids)

		// Print information about each file.
		fmt.Printf("  Files:\n")
		for _, id := range ids {
			func() {
				fmt.Printf("    %s:\n", id)

				// Resolve the file reference.
				ref, err := dep.Resources.FileSystem.ResolveFile(id)
				if err != nil {
					fmt.Printf("      Path:     %v\n", err)
					return
				}

				// Generate a file path.
				path, err := ref.Path()
				if err != nil {
					fmt.Printf("      Path:     %v\n", err)
					return
				}
				fmt.Printf("      Path:     %s\n", path)

				// Attempt to open the parent directory.
				dir, err := localfs.OpenDir(ref.Dir())
				if err != nil {
					if os.IsNotExist(err) {
						fmt.Printf("      Status:   Missing\n")
					} else {
						fmt.Printf("      Status:   %v\n", err)
					}
					return
				}
				defer dir.Close()

				// Stat the file path.
				fi, err := dir.System().Stat(ref.FilePath)
				if err != nil {
					if os.IsNotExist(err) {
						fmt.Printf("      Status:   Missing\n")
					} else {
						fmt.Printf("      Status:   %v\n", err)
					}
					return
				}

				// Make sure it's a regular file.
				if !fi.Mode().IsRegular() {
					fmt.Printf("      Status:   Not A File\n")
					return
				}

				// Report statistics.
				fmt.Printf("      Status:   Present\n")
				fmt.Printf("      Modified: %s\n", fi.ModTime())
				fmt.Printf("      Size:     %d bytes(s)\n", fi.Size())
			}()
		}
	}

	return nil
}

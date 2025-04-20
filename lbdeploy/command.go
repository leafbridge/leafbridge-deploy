package lbdeploy

// CommandType identifies the type of a command.
type CommandType string

// Command types.
const (
	CommandTypeExe          = "exe"
	CommandTypeMSIInstall   = "msi-install"
	CommandTypeMSIUninstall = "msi-uninstall"
)

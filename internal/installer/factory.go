package installer

import "os"

// Compile-time interface checks for Installer implementations.
var (
	_ Installer = (*OpenCodeInstaller)(nil)
	_ Installer = (*EngramInstaller)(nil)
	_ Installer = (*ContextModeInstaller)(nil)
	_ Installer = (*PlaywrightInstaller)(nil)
)

// NewDefaultRegistry creates a Registry with all supported installers.
func NewDefaultRegistry(homeDir string) (*Registry, error) {
	cmd := &OSCommandRunner{}
	return NewRegistry(
		&OpenCodeInstaller{Cmd: cmd, HomeDir: homeDir, StatPath: os.Stat},
		&EngramInstaller{Cmd: cmd, HomeDir: homeDir, RemoveFunc: os.Remove},
		&ContextModeInstaller{Cmd: cmd},
		&PlaywrightInstaller{Cmd: cmd, HomeDir: homeDir, StatPath: os.Stat},
	)
}

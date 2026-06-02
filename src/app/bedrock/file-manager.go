package bedrock

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/snivilised/jaywalk/src/agenor/core"
)

// FileManager provides a single point of reference for all XDG-compliant
// paths used by jay. It consolidates every environment variable and file
// location the application cares about.
//
// Environment variables respected:
//
//	JAY_CONFIG_DIR   - overrides the config directory
//	                  (default: $XDG_CONFIG_HOME/jay → ~/.config/jay)
//	JAY_STATE_DIR    - overrides the state directory
//	                  (default: $XDG_STATE_HOME/jay → ~/.local/state/jay)
//	JAY_CACHE_DIR    - overrides the cache directory
//	                  (default: $XDG_CACHE_HOME/jay → ~/.cache/jay)
//	JAY_THEMES_DIR   - overrides the themes directory
//	                  (default: $XDG_CONFIG_HOME/jay/themes → ~/.config/jay/themes)
//
// Standard XDG variables that are also respected:
//
//	XDG_CONFIG_HOME  - base config directory     (default: ~/.config)
//	XDG_STATE_HOME   - base state directory      (default: ~/.local/state)
//	XDG_CACHE_HOME   - base cache directory      (default: ~/.cache)
//
// Files managed by jay:
//
//	Config file:  <configHome>/jay.yaml (also jay.yml, jay.json, jay.toml)
//	Themes dir:   <configHome>/themes/
//	Admin dir:    <stateHome>/admin/
//	Log file:     <adminHome>/logs/jay.log
//	Resume state: <adminHome>/resume/
//
// Resolution priority for each directory is:
//  1. Jay-specific env var override
//  2. XDG env var
//  3. Compiled default
type FileManager struct {
	configHome string
	stateHome  string
	cacheHome  string
	themesDir  string
}

// NewFileManager constructs a FileManager with paths resolved from
// environment variables with XDG fallbacks.
func NewFileManager() *FileManager {
	fm := &FileManager{
		configHome: filepath.Join(xdgConfigHome(), AppName),
		stateHome:  filepath.Join(xdgStateHome(), AppName),
		cacheHome:  filepath.Join(xdgCacheHome(), AppName),
	}

	// Override with jay-specific env vars if set
	if env := core.Getenv("JAY_CONFIG_DIR"); env != "" {
		fm.configHome = fm.resolvePath(env)
	}
	if env := core.Getenv("JAY_STATE_DIR"); env != "" {
		fm.stateHome = fm.resolvePath(env)
	}
	if env := core.Getenv("JAY_CACHE_DIR"); env != "" {
		fm.cacheHome = fm.resolvePath(env)
	}

	// Themes dir
	if env := core.Getenv("JAY_THEMES_DIR"); env != "" {
		fm.themesDir = fm.resolvePath(env)
	} else {
		fm.themesDir = filepath.Join(fm.configHome, "themes")
	}

	return fm
}

// xdgConfigHome returns the XDG config home directory.
// Priority: XDG_CONFIG_HOME env var → ~/.config
func xdgConfigHome() string {
	if env := core.Getenv("XDG_CONFIG_HOME"); env != "" {
		return env
	}
	home, err := core.Home()
	if err != nil {
		return filepath.Join(".config")
	}
	return filepath.Join(home, ".config")
}

// xdgStateHome returns the XDG state home directory.
// Priority: XDG_STATE_HOME env var → ~/.local/state
func xdgStateHome() string {
	if env := core.Getenv("XDG_STATE_HOME"); env != "" {
		return env
	}
	home, err := core.Home()
	if err != nil {
		return filepath.Join(".local", "state")
	}
	return filepath.Join(home, ".local", "state")
}

// xdgCacheHome returns the XDG cache home directory.
// Priority: XDG_CACHE_HOME env var → ~/.cache
func xdgCacheHome() string {
	if env := core.Getenv("XDG_CACHE_HOME"); env != "" {
		return env
	}
	home, err := core.Home()
	if err != nil {
		return filepath.Join(".cache")
	}
	return filepath.Join(home, ".cache")
}

// ConfigHome returns the config directory path.
func (fm *FileManager) ConfigHome() string {
	return fm.configHome
}

// StateHome returns the state directory path.
func (fm *FileManager) StateHome() string {
	return fm.stateHome
}

// CacheHome returns the cache directory path.
func (fm *FileManager) CacheHome() string {
	return fm.cacheHome
}

// ThemesDir returns the themes directory path.
func (fm *FileManager) ThemesDir() string {
	return fm.themesDir
}

// AdminPath returns the admin base directory path.
func (fm *FileManager) AdminPath() string {
	return filepath.Join(fm.stateHome, "admin")
}

// LogPath returns the path for the log file, derived from the admin path.
func (fm *FileManager) LogPath() string {
	return filepath.Join(fm.AdminPath(), "logs", "jay.log")
}

// ResolvePath expands ~ and environment variables in the given path.
func (fm *FileManager) ResolvePath(p string) string {
	return resolvePath(p)
}

// resolvePath is the internal implementation used during construction.
func (fm *FileManager) resolvePath(p string) string {
	return resolvePath(p)
}

// resolvePath expands ~ to the user's home directory and then expands
// environment variables ($VAR or ${VAR}) in the given path.
func resolvePath(p string) string {
	result := p
	if strings.HasPrefix(result, "~") {
		home, err := core.Home()
		if err == nil {
			result = filepath.Join(home, result[1:])
		}
	}
	result = os.ExpandEnv(result)
	return result
}

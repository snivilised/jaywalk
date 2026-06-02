package bedrock

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	nef "github.com/snivilised/nefilim"
	"github.com/spf13/viper"
)

// ViewConfigLoader resolves and loads a named view's config section
// from jay.ui.yml. Constructed once by Bootstrap and shared across
// the application lifetime.
type ViewConfigLoader struct {
	configHome string
	fS         nef.UniversalFS // non-nil when using a mocked FS (e.g. luna.MemFS in tests)
}

// NewViewConfigLoader constructs a ViewConfigLoader. The configHome is
// the directory where jay.ui.yml is expected (typically ~/.config/jay).
func NewViewConfigLoader(configHome string) *ViewConfigLoader {
	return &ViewConfigLoader{
		configHome: configHome,
	}
}

// NewViewConfigLoaderWithFS constructs a ViewConfigLoader backed by the
// given filesystem. Passing a luna.MemFS lets tests avoid touching the
// real filesystem.
func NewViewConfigLoaderWithFS(configHome string, fS nef.UniversalFS) *ViewConfigLoader {
	return &ViewConfigLoader{
		configHome: configHome,
		fS:         fS,
	}
}

// viewExtensions lists the file extensions tried when loading view
// config, in priority order.
var viewExtensions = []string{".yaml", ".yml"}

// readFile reads the config file at the given path. When a mocked FS is
// set (l.fS != nil), it reads from that FS; otherwise it delegates to
// Viper's file-based ReadInConfig.
func (l *ViewConfigLoader) readFile(v *viper.Viper, path string) error {
	if l.fS != nil {
		data, err := l.fS.ReadFile(path)
		if err != nil {
			return err
		}
		return v.ReadConfig(bytes.NewReader(data))
	}
	v.SetConfigFile(path)
	return v.ReadInConfig()
}

// fileExists checks if a config file exists at the given path. When a
// mocked FS is set it queries that FS; otherwise it stats the real disk.
func (l *ViewConfigLoader) fileExists(path string) bool {
	if l.fS != nil {
		_, err := l.fS.Stat(path)
		return err == nil
	}
	_, err := os.Stat(path)
	return err == nil
}

// ResolvePath returns the full path of the first existing view config
// file (jay.ui.yaml or jay.ui.yml) under configHome. Returns empty
// string when no file is found.
func (l *ViewConfigLoader) ResolvePath() string {
	for _, ext := range viewExtensions {
		path := filepath.Join(l.configHome, "jay.ui"+ext)
		if l.fileExists(path) {
			return path
		}
	}
	return ""
}

// Load decodes the ui.<viewName> section from jay.ui.yml into dest.
// Returns nil when the file or section is absent (caller should use
// defaults). Returns an error only when a found file cannot be decoded.
// Adding a new view is just: loader.Load("view-name", &cfg).
func (l *ViewConfigLoader) Load(viewName string, dest any) error {
	for _, ext := range viewExtensions {
		path := l.configHome + "/jay.ui" + ext

		v := viper.New()
		v.SetConfigType("yaml")

		if err := l.readFile(v, path); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("reading jay.ui%s: %w", ext, err)
		}

		key := "ui." + viewName
		raw := v.Sub(key)
		if raw == nil {
			return nil
		}

		if err := raw.Unmarshal(dest, mapstructureTagOption()); err != nil {
			return fmt.Errorf("decoding %s: %w", key, err)
		}

		return nil
	}

	// No file found at any extension - not an error, caller uses defaults.
	return nil
}

package bedrock

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"

	nef "github.com/snivilised/nefilim"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism"
)

const (
	// ThemesDirEnvVar is the environment variable that overrides the
	// default XDG themes directory. When set, jay loads themes from
	// this path instead of ~/.config/jay/themes/.
	ThemesDirEnvVar = "JAY_THEMES_DIR"

	// ThemeSystemName is the reserved name for the built-in ANSI-16
	// palette that respects the user's terminal theme. Selecting this
	// name bypasses file loading entirely.
	ThemeSystemName = "system"
)

// ThemeLoader resolves and loads a named theme from the themes directory.
// Constructed once by Bootstrap and shared across the application lifetime.
type ThemeLoader struct {
	themesDir string
	fS        nef.UniversalFS // non-nil when using a mocked FS (e.g. luna.MemFS in tests)
}

// NewThemeLoader constructs a ThemeLoader. The themes directory is
// resolved from JAY_THEMES_DIR if set (with ~ expansion),
// otherwise from the XDG config base at ~/.config/jay/themes/.
func NewThemeLoader() *ThemeLoader {
	dir := core.Getenv(ThemesDirEnvVar)

	if dir == "" {
		dir = filepath.Join(xdg.ConfigHome, AppName, "themes")
	} else {
		// Expand ~ in env var value
		if len(dir) > 0 && dir[0] == '~' {
			if home, err := core.Home(); err == nil {
				dir = filepath.Join(home, dir[1:])
			}
		}
	}

	return &ThemeLoader{
		themesDir: dir,
	}
}

// NewThemeLoaderWithDir constructs a ThemeLoader with an explicitly
// resolved themes directory path. This is the preferred constructor
// when the FileManager has already resolved paths (e.g., from
// bootstrap). The zero-arg NewThemeLoader is retained for backward
// compatibility with tests.
func NewThemeLoaderWithDir(dir string) *ThemeLoader {
	return &ThemeLoader{
		themesDir: dir,
	}
}

// NewThemeLoaderWithFS constructs a ThemeLoader backed by the given
// filesystem. Passing a luna.MemFS lets tests avoid touching the
// real filesystem.
func NewThemeLoaderWithFS(dir string, fS nef.UniversalFS) *ThemeLoader {
	return &ThemeLoader{
		themesDir: dir,
		fS:        fS,
	}
}

// themeExtensions lists the file extensions tried when loading a
// theme file, in priority order.
var themeExtensions = []string{".yaml", ".yml"}

// readFile reads the theme file at the given path. When a mocked FS is
// set (tl.fS != nil), it reads from that FS; otherwise it delegates to
// Viper's file-based ReadInConfig.
func (tl *ThemeLoader) readFile(v *viper.Viper, path string) error {
	if tl.fS != nil {
		data, err := tl.fS.ReadFile(path)
		if err != nil {
			return err
		}
		return v.ReadConfig(bytes.NewReader(data))
	}
	v.SetConfigFile(path)
	return v.ReadInConfig()
}

// Load returns the prism.Palette for the named theme. The name
// "system" returns SystemPalette() without reading any file.
// Any other name is resolved to <themesDir>/<name>.<ext> for each
// registered extension and decoded via mapstructure. Returns an
// error if no file is found or the file cannot be decoded.
func (tl *ThemeLoader) Load(name string) (prism.Palette, error) {
	if name == "" || name == ThemeSystemName {
		return prism.SystemPalette(), nil
	}

	var lastErr error
	for _, ext := range themeExtensions {
		path := filepath.Join(tl.themesDir, name+ext)

		v := viper.New()
		v.SetConfigType("yaml")

		if err := tl.readFile(v, path); err != nil {
			if os.IsNotExist(err) {
				lastErr = err
				continue
			}

			return prism.Palette{}, fmt.Errorf(
				"reading theme %q at %s: %w",
				name,
				path,
				err,
			)
		}

		raw := v.Sub("palette")
		if raw == nil {
			return prism.Palette{}, fmt.Errorf(
				"theme %q: missing required 'palette' key in %s",
				name,
				path,
			)
		}

		var palette prism.Palette

		if err := raw.Unmarshal(&palette, mapstructureTagOption()); err != nil {
			return prism.Palette{}, fmt.Errorf(
				"theme %q: decoding palette from %s: %w",
				name,
				path,
				err,
			)
		}

		return palette, nil
	}

	// None of the extensions matched — report the first expected path
	// for clarity.
	first := filepath.Join(tl.themesDir, name+themeExtensions[0])
	if os.IsNotExist(lastErr) {
		return prism.Palette{}, fmt.Errorf(
			"theme %q not found — tried %s/.%s{.yaml,.yml}",
			name,
			tl.themesDir,
			name,
		)
	}

	return prism.Palette{}, fmt.Errorf(
		"theme %q not found - expected file at %s",
		name,
		first,
	)
}

// NameFromPalette extracts the component's gradient name or empty string.
func (tl *ThemeLoader) NameFromPalette(p prism.Palette, componentName string) string {
	components := p.Highlights.Components
	if len(components) == 0 {
		return ""
	}
	name, ok := components[componentName]
	if !ok {
		return ""
	}
	return name
}

// ThemesDir returns the resolved themes directory path. Useful for
// error messages and the --theme flag's help text.
func (tl *ThemeLoader) ThemesDir() string {
	return tl.themesDir
}

// mapstructureTagOption returns a viper.DecoderConfigOption that
// instructs mapstructure to use the "mapstructure" struct tag when
// decoding YAML keys into Go field names. This enables kebab-case YAML
// keys to map correctly into the palette structs via their
// mapstructure tags.
//
// The explicit viper.DecoderConfigOption cast is required because Go
// does not implicitly convert a func literal to a named function type
// even when the underlying signatures are identical.
func mapstructureTagOption() viper.DecoderConfigOption {
	return viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "mapstructure"
	})
}

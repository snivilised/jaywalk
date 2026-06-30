package bedrock

import (
	"fmt"

	nef "github.com/snivilised/nefilim"
	"gopkg.in/yaml.v3"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// themeFile is the top-level envelope written to disk. It matches
// the structure ThemeLoader.Load expects to read back.
type themeFile struct {
	Palette contract.Palette `yaml:"palette"`
}

// ThemeWriter serialises a contract.Palette to a YAML theme file.
// Symmetric counterpart to ThemeLoader. Uses gopkg.in/yaml.v3
// directly; Viper is not involved in the write path.
type ThemeWriter struct {
	themesDir string
	fS        nef.UniversalFS
}

// NewThemeWriter constructs a ThemeWriter that writes to the real
// filesystem under themesDir.
func NewThemeWriter(themesDir string) *ThemeWriter {
	return &ThemeWriter{
		themesDir: themesDir,
	}
}

// NewThemeWriterWithFS constructs a ThemeWriter backed by the given
// filesystem. Passing a luna.NewMemFS() lets tests avoid touching
// the real filesystem.
func NewThemeWriterWithFS(themesDir string, fS nef.UniversalFS) *ThemeWriter {
	return &ThemeWriter{
		themesDir: themesDir,
		fS:        fS,
	}
}

// Write encodes palette as a YAML theme file at
// <themesDir>/<name>.yaml. The palette is wrapped under the
// top-level "palette:" key so the output is readable by
// ThemeLoader.Load. Uses atomicWriteFile for safe cross-platform
// writes; no partial file is visible to readers on success.
func (tw *ThemeWriter) Write(name string, palette contract.Palette) error {
	envelope := themeFile{Palette: palette}

	data, err := yaml.Marshal(&envelope)
	if err != nil {
		return fmt.Errorf("marshalling theme %q: %w", name, err)
	}

	path := tw.themesDir + "/" + name + ".yaml"

	if err := atomicWriteFile(tw.fS, path, data); err != nil {
		return fmt.Errorf("writing theme %q to %s: %w", name, path, err)
	}

	return nil
}

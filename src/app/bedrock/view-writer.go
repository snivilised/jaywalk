package bedrock

import (
	"fmt"

	nef "github.com/snivilised/nefilim"
	"gopkg.in/yaml.v3"
)

// viewConfigFile is the top-level envelope written to disk. It matches
// the structure ViewConfigLoader expects to read back.
type viewConfigFile struct {
	UI FullViewConfig `yaml:"ui"`
}

// ViewConfigWriter serialises a FullViewConfig to jay.ui.yaml.
// Uses gopkg.in/yaml.v3 directly; Viper is not involved in the
// write path. Uses atomicWriteFile for safe cross-platform writes.
type ViewConfigWriter struct {
	configHome string
	fS         nef.UniversalFS
}

// NewViewConfigWriter constructs a ViewConfigWriter that writes to
// the real filesystem under configHome.
func NewViewConfigWriter(configHome string) *ViewConfigWriter {
	return &ViewConfigWriter{
		configHome: configHome,
	}
}

// NewViewConfigWriterWithFS constructs a ViewConfigWriter backed by
// the given filesystem. Passing a luna.NewMemFS() lets tests avoid
// touching the real filesystem.
func NewViewConfigWriterWithFS(configHome string, fS nef.UniversalFS) *ViewConfigWriter {
	return &ViewConfigWriter{
		configHome: configHome,
		fS:         fS,
	}
}

// Write serialises cfg to <configHome>/jay.ui.yaml using
// atomicWriteFile for safe cross-platform writes.
func (w *ViewConfigWriter) Write(cfg FullViewConfig) error {
	envelope := viewConfigFile{UI: cfg}

	data, err := yaml.Marshal(&envelope)
	if err != nil {
		return fmt.Errorf("marshalling view config: %w", err)
	}

	path := w.configHome + "/jay.ui.yaml"

	if err := atomicWriteFile(w.fS, path, data); err != nil {
		return fmt.Errorf("writing view config to %s: %w", path, err)
	}

	return nil
}

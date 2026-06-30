package bedrock

import "github.com/snivilised/jaywalk/src/prism/contract"

// FullViewConfig is the concrete on-disk representation of jay.ui.yml.
// It replaces the former sealed ViewConfig interface with a single
// struct that carries all view sections. The ViewConfigWriter
// serialises this struct; the ViewConfigLoader populates it.
//
// The struct mirrors the YAML shape:
//
//	ui:
//	  linear:   { ... }
//	  highway:  { ... }
//	  porthole: { ... }
type FullViewConfig struct {
	Linear   LinearConfig  `yaml:"linear,omitempty"`
	Highway  HighwayConfig `yaml:"highway,omitempty"`
	Porthole HighwayConfig `yaml:"porthole,omitempty"`
}

// LoadConfig reads all view sections from the loader into a
// FullViewConfig. The palette is used to derive per-palette values
// such as the highway animation gradient. Returns the concrete
// on-disk config that the ui package's New accepts directly.
func LoadConfig(loader *ViewConfigLoader, _ contract.Palette) (*FullViewConfig, error) {
	cfg := &FullViewConfig{}

	if loader != nil {
		if err := loader.Load("linear", &cfg.Linear); err != nil {
			return nil, err
		}

		if err := loader.Load("highway", &cfg.Highway); err != nil {
			return nil, err
		}

		if err := loader.Load("porthole", &cfg.Porthole); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

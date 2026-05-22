package bedrock_test

import (
	"os"
	"path/filepath"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/prism"
	"github.com/snivilised/nefilim/test/luna"
)

// validThemeYAML is a minimal well-formed theme file covering a subset
// of palette entries. The loader must tolerate absent entries - they
// will be zero-value SemanticColour structs, which resolve without error.
const validThemeYAML = `
palette:
  banner:
    ansi16: "magenta"
    ansi256: "99"
    true-color: "#5C4AE4"
  directory:
    ansi16: "cyan"
    ansi256: "116"
    true-color: "#89DCEB"
  error:
    ansi16: "red"
    ansi256: "211"
    true-color: "#F38BA8"
  tree-icons:
    root-icon: "✻"
    directory-icon: "📁"
    file-icon: "🔖"
    skipped-icon: "⛔️"
    elapsed-icon: "⏰"
    branch-vertical: "│"
    branch-joint: "├── "
    branch-last: "└── "
    branch-indent: "  "
`

// invalidANSI16YAML contains a palette entry with an unrecognised
// ansi16 name. The loader itself does not validate colour names - that
// is prism's responsibility at NewTheme time. So loading succeeds but
// the returned palette will fail at NewTheme.
const invalidANSI16YAML = `
palette:
  directory:
    ansi16: "notacolour"
`

// missingPaletteYAML is a valid YAML file that lacks the required
// top-level 'palette' key.
const missingPaletteYAML = `
name: broken-theme
`

var _ = Describe("ThemeLoader", Ordered, func() {
	var themesDir string

	BeforeAll(func() {
		// Create a temporary themes directory for env-var resolution tests
		// that still use the real filesystem.
		var err error
		themesDir, err = os.MkdirTemp("", "jay-themes-*")
		Expect(err).To(BeNil())
	})

	AfterAll(func() {
		_ = os.RemoveAll(themesDir)
	})

	// ------------------------------------------------------------------
	// JAY_THEMES_DIR env var
	// ------------------------------------------------------------------

	Describe("NewThemeLoader", func() {
		Context("when JAY_THEMES_DIR is set", func() {
			It("uses the env var path as the themes directory", func() {
				DeferCleanup(func() {
					_ = os.Unsetenv(bedrock.ThemesDirEnvVar)
				})

				_ = os.Setenv(bedrock.ThemesDirEnvVar, themesDir)
				loader := bedrock.NewThemeLoader()

				Expect(loader.ThemesDir()).To(Equal(themesDir))
			})
		})

		Context("when JAY_THEMES_DIR is not set", func() {
			It("uses the XDG default path containing 'jay/themes'", func() {
				DeferCleanup(func() {
					_ = os.Unsetenv(bedrock.ThemesDirEnvVar)
				})

				_ = os.Unsetenv(bedrock.ThemesDirEnvVar)
				loader := bedrock.NewThemeLoader()

				Expect(loader.ThemesDir()).To(ContainSubstring(
					filepath.Join(bedrock.AppName, "themes"),
				))
			})
		})
	})

	// ------------------------------------------------------------------
	// Load - system palette
	// ------------------------------------------------------------------

	Describe("Load", func() {
		Context("when the theme name is empty", func() {
			It("returns the system palette without reading any file", func() {
				loader := bedrock.NewThemeLoader()

				palette, err := loader.Load("")

				Expect(err).To(BeNil())
				Expect(palette).To(Equal(prism.SystemPalette()))
			})
		})

		Context("when the theme name is 'system'", func() {
			It("returns the system palette without reading any file", func() {
				loader := bedrock.NewThemeLoader()

				palette, err := loader.Load(bedrock.ThemeSystemName)

				Expect(err).To(BeNil())
				Expect(palette).To(Equal(prism.SystemPalette()))
			})
		})

		// ------------------------------------------------------------------
		// Load - file-based themes (MemFS)
		// ------------------------------------------------------------------

		Context("when loading from a valid theme file", func() {
			It("decodes the palette correctly", func() {
				fS := luna.NewMemFS()
				fS.MapFS["themes/my-theme.yaml"] = &fstest.MapFile{
					Data: []byte(validThemeYAML),
				}

				loader := bedrock.NewThemeLoaderWithFS("themes", fS)
				palette, err := loader.Load("my-theme")

				Expect(err).To(BeNil())
				Expect(palette.Directory.ANSI16).To(Equal("cyan"))
				Expect(palette.Error.ANSI16).To(Equal("red"))
				Expect(palette.TreeIcons["root-icon"]).To(Equal("✻"))
				Expect(palette.TreeIcons["directory-icon"]).To(Equal("📁"))
				Expect(palette.TreeIcons["file-icon"]).To(Equal("🔖"))
				Expect(palette.TreeIcons["skipped-icon"]).To(Equal("⛔️"))
				Expect(palette.TreeIcons["elapsed-icon"]).To(Equal("⏰"))
			})
		})

		Context("when the theme file does not exist", func() {
			It("returns an error mentioning the theme name and tried paths", func() {
				fS := luna.NewMemFS()
				loader := bedrock.NewThemeLoaderWithFS("themes", fS)

				_, err := loader.Load("nonexistent-theme")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("nonexistent-theme"))
				Expect(err.Error()).To(ContainSubstring(".yaml,.yml"))
			})
		})

		Context("when the theme file has a .yml extension", func() {
			It("loads successfully", func() {
				fS := luna.NewMemFS()
				fS.MapFS["themes/dot-yml.yml"] = &fstest.MapFile{
					Data: []byte(validThemeYAML),
				}

				loader := bedrock.NewThemeLoaderWithFS("themes", fS)
				palette, err := loader.Load("dot-yml")

				Expect(err).To(BeNil())
				Expect(palette.Directory.ANSI16).To(Equal("cyan"))
			})
		})

		Context("when the theme file is missing the palette key", func() {
			It("returns an error mentioning 'palette'", func() {
				fS := luna.NewMemFS()
				fS.MapFS["themes/no-palette.yaml"] = &fstest.MapFile{
					Data: []byte(missingPaletteYAML),
				}

				loader := bedrock.NewThemeLoaderWithFS("themes", fS)
				_, err := loader.Load("no-palette")

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("palette"))
			})
		})

		Context("when a theme file has an unrecognised ansi16 name", func() {
			It("loads successfully - colour validation is prism's responsibility", func() {
				fS := luna.NewMemFS()
				fS.MapFS["themes/bad-colours.yaml"] = &fstest.MapFile{
					Data: []byte(invalidANSI16YAML),
				}

				loader := bedrock.NewThemeLoaderWithFS("themes", fS)
				palette, err := loader.Load("bad-colours")

				// Loader does not validate colour names - it returns the raw
				// decoded palette. The error surfaces later in prism.NewTheme.
				Expect(err).To(BeNil())
				Expect(palette.Directory.ANSI16).To(Equal("notacolour"))
			})
		})
	})
})

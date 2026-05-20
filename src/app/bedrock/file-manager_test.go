package bedrock_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/app/bedrock"
)

var _ = Describe("FileManager", func() {
	var originalEnvVars map[string]string

	saveAndClear := func(vars ...string) {
		originalEnvVars = make(map[string]string, len(vars))
		for _, v := range vars {
			originalEnvVars[v] = os.Getenv(v)
			_ = os.Unsetenv(v)
		}
	}

	restoreEnv := func() {
		for k, v := range originalEnvVars {
			if v == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, v)
			}
		}
	}

	BeforeEach(func() {
		saveAndClear(
			"JAY_CONFIG_DIR", "JAY_STATE_DIR", "JAY_CACHE_DIR", "JAY_THEMES_DIR",
			"XDG_CONFIG_HOME", "XDG_STATE_HOME", "XDG_CACHE_HOME",
		)
	})

	AfterEach(func() {
		restoreEnv()
	})

	// ------------------------------------------------------------------
	// Default paths (no env overrides)
	// ------------------------------------------------------------------

	Describe("NewFileManager defaults", func() {
		It("ConfigHome contains 'jay' in the path", func() {
			fm := bedrock.NewFileManager()
			Expect(fm.ConfigHome()).To(ContainSubstring(bedrock.AppName))
		})

		It("StateHome contains 'jay' in the path", func() {
			fm := bedrock.NewFileManager()
			Expect(fm.StateHome()).To(ContainSubstring(bedrock.AppName))
		})

		It("CacheHome contains 'jay' in the path", func() {
			fm := bedrock.NewFileManager()
			Expect(fm.CacheHome()).To(ContainSubstring(bedrock.AppName))
		})

		It("ThemesDir is ConfigHome + '/themes'", func() {
			fm := bedrock.NewFileManager()
			Expect(fm.ThemesDir()).To(Equal(filepath.Join(fm.ConfigHome(), "themes")))
		})

		It("AdminPath is StateHome + '/admin'", func() {
			fm := bedrock.NewFileManager()
			Expect(fm.AdminPath()).To(Equal(filepath.Join(fm.StateHome(), "admin")))
		})

		It("LogPath is AdminPath + '/logs/jay.log'", func() {
			fm := bedrock.NewFileManager()
			Expect(fm.LogPath()).To(Equal(filepath.Join(fm.AdminPath(), "logs", "jay.log")))
		})
	})

	// ------------------------------------------------------------------
	// Overrides via JAY_* env vars
	// ------------------------------------------------------------------

	Describe("JAY_CONFIG_DIR override", func() {
		It("uses the env var value for ConfigHome", func() {
			_ = os.Setenv("JAY_CONFIG_DIR", "/custom/config/jay")
			fm := bedrock.NewFileManager()
			Expect(fm.ConfigHome()).To(Equal("/custom/config/jay"))
		})

		It("expands ~ in the env var value", func() {
			home, err := os.UserHomeDir()
			Expect(err).To(BeNil())
			_ = os.Setenv("JAY_CONFIG_DIR", "~/myconfig")
			fm := bedrock.NewFileManager()
			Expect(fm.ConfigHome()).To(Equal(filepath.Join(home, "myconfig")))
		})

		It("expands $HOME in the env var value", func() {
			home := os.Getenv("HOME")
			_ = os.Setenv("JAY_CONFIG_DIR", "$HOME/customcfg")
			fm := bedrock.NewFileManager()
			Expect(fm.ConfigHome()).To(Equal(filepath.Join(home, "customcfg")))
		})
	})

	Describe("JAY_STATE_DIR override", func() {
		It("uses the env var value for StateHome", func() {
			_ = os.Setenv("JAY_STATE_DIR", "/custom/state/jay")
			fm := bedrock.NewFileManager()
			Expect(fm.StateHome()).To(Equal("/custom/state/jay"))
		})

		It("AdminPath and LogPath are relative to the overridden StateHome", func() {
			_ = os.Setenv("JAY_STATE_DIR", "/alt/state")
			fm := bedrock.NewFileManager()
			Expect(fm.AdminPath()).To(Equal(filepath.Join("/alt/state", "admin")))
			Expect(fm.LogPath()).To(Equal(filepath.Join("/alt/state", "admin", "logs", "jay.log")))
		})
	})

	Describe("JAY_CACHE_DIR override", func() {
		It("uses the env var value for CacheHome", func() {
			_ = os.Setenv("JAY_CACHE_DIR", "/custom/cache")
			fm := bedrock.NewFileManager()
			Expect(fm.CacheHome()).To(Equal("/custom/cache"))
		})
	})

	Describe("JAY_THEMES_DIR override", func() {
		It("uses the env var value for ThemesDir", func() {
			_ = os.Setenv("JAY_THEMES_DIR", "/custom/themes")
			fm := bedrock.NewFileManager()
			Expect(fm.ThemesDir()).To(Equal("/custom/themes"))
		})

		It("expands ~ in the env var value", func() {
			home, err := os.UserHomeDir()
			Expect(err).To(BeNil())
			_ = os.Setenv("JAY_THEMES_DIR", "~/mythemes")
			fm := bedrock.NewFileManager()
			Expect(fm.ThemesDir()).To(Equal(filepath.Join(home, "mythemes")))
		})

		It("is independent of ConfigHome when set", func() {
			_ = os.Setenv("JAY_CONFIG_DIR", "/cfg")
			_ = os.Setenv("JAY_THEMES_DIR", "/thm")
			fm := bedrock.NewFileManager()
			Expect(fm.ThemesDir()).To(Equal("/thm"))
			Expect(fm.ThemesDir()).NotTo(ContainSubstring("/cfg"))
		})
	})

	// ------------------------------------------------------------------
	// XDG env var overrides (standard)
	// ------------------------------------------------------------------

	Describe("XDG_CONFIG_HOME affects default ConfigHome", func() {
		It("uses XDG_CONFIG_HOME/jay when XDG_CONFIG_HOME is set", func() {
			_ = os.Setenv("XDG_CONFIG_HOME", "/xdg/cfg")
			fm := bedrock.NewFileManager()
			Expect(fm.ConfigHome()).To(Equal(filepath.Join("/xdg/cfg", bedrock.AppName)))
		})

		It("is overridden by JAY_CONFIG_DIR", func() {
			_ = os.Setenv("XDG_CONFIG_HOME", "/xdg/cfg")
			_ = os.Setenv("JAY_CONFIG_DIR", "/override/jay")
			fm := bedrock.NewFileManager()
			Expect(fm.ConfigHome()).To(Equal("/override/jay"))
		})
	})

	Describe("XDG_STATE_HOME affects default StateHome", func() {
		It("uses XDG_STATE_HOME/jay when XDG_STATE_HOME is set", func() {
			_ = os.Setenv("XDG_STATE_HOME", "/xdg/state")
			fm := bedrock.NewFileManager()
			Expect(fm.StateHome()).To(Equal(filepath.Join("/xdg/state", bedrock.AppName)))
		})
	})

	Describe("XDG_CACHE_HOME affects default CacheHome", func() {
		It("uses XDG_CACHE_HOME/jay when XDG_CACHE_HOME is set", func() {
			_ = os.Setenv("XDG_CACHE_HOME", "/xdg/cache")
			fm := bedrock.NewFileManager()
			Expect(fm.CacheHome()).To(Equal(filepath.Join("/xdg/cache", bedrock.AppName)))
		})
	})

	// ------------------------------------------------------------------
	// ResolvePath
	// ------------------------------------------------------------------

	Describe("ResolvePath", func() {
		It("expands ~ to the user's home directory", func() {
			fm := bedrock.NewFileManager()
			home, err := os.UserHomeDir()
			Expect(err).To(BeNil())
			result := fm.ResolvePath("~/some/path")
			Expect(result).To(Equal(filepath.Join(home, "some", "path")))
		})

		It("expands $HOME to the user's home directory", func() {
			fm := bedrock.NewFileManager()
			home := os.Getenv("HOME")
			result := fm.ResolvePath("$HOME/some/path")
			Expect(result).To(Equal(filepath.Join(home, "some", "path")))
		})

		It("passes through already-absolute paths unchanged (minus expansion)", func() {
			fm := bedrock.NewFileManager()
			result := fm.ResolvePath("/absolute/path")
			Expect(result).To(Equal("/absolute/path"))
		})

		It("expands $XDG_CONFIG_HOME in a path", func() {
			_ = os.Setenv("XDG_CONFIG_HOME", "/custom/xdg")
			fm := bedrock.NewFileManager()
			result := fm.ResolvePath("$XDG_CONFIG_HOME/jay/config.yaml")
			Expect(result).To(Equal("/custom/xdg/jay/config.yaml"))
		})
	})

	// ------------------------------------------------------------------
	// XDG defaults for state/cache when those env vars are unset
	// ------------------------------------------------------------------

	Describe("XDG fallback defaults", func() {
		It("StateHome falls back to ~/.local/state/jay when XDG_STATE_HOME and JAY_STATE_DIR are unset", func() {
			_ = os.Unsetenv("XDG_STATE_HOME")
			_ = os.Unsetenv("JAY_STATE_DIR")
			fm := bedrock.NewFileManager()
			Expect(fm.StateHome()).To(Or(
				ContainSubstring(filepath.Join(".local", "state", bedrock.AppName)),
				ContainSubstring(filepath.Join(bedrock.AppName)),
			))
		})

		It("CacheHome falls back to ~/.cache/jay when XDG_CACHE_HOME and JAY_CACHE_DIR are unset", func() {
			_ = os.Unsetenv("XDG_CACHE_HOME")
			_ = os.Unsetenv("JAY_CACHE_DIR")
			fm := bedrock.NewFileManager()
			Expect(fm.CacheHome()).To(ContainSubstring(filepath.Join("cache", bedrock.AppName)))
		})
	})
})

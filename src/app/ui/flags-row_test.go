package ui

import (
	"os"
	"testing/fstest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/nefilim/test/luna"

	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("normaliseFlagsRowPosition", func() {
	It("returns the default for an empty string", func() {
		Expect(normaliseFlagsRowPosition("")).To(Equal(contract.PositionBottom))
	})

	It("returns top unchanged", func() {
		Expect(normaliseFlagsRowPosition(contract.PositionTop)).To(Equal(contract.PositionTop))
	})

	It("returns bottom unchanged", func() {
		Expect(normaliseFlagsRowPosition(contract.PositionBottom)).To(Equal(contract.PositionBottom))
	})

	It("returns the default and writes a warning for an unrecognised value", func() {
		// Capture stderr for the duration of the call. Using t.TempDir
		// isn't available here (this is a Ginkgo test), so we manually
		// manage the temp dir.
		tmpDir, err := os.MkdirTemp("", "flags-row-test-*")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = os.RemoveAll(tmpDir) }()

		originalStderr := os.Stderr
		r, w, pipeErr := os.Pipe()
		Expect(pipeErr).NotTo(HaveOccurred())
		os.Stderr = w
		defer func() { os.Stderr = originalStderr }()

		got := normaliseFlagsRowPosition("sideways")
		_ = w.Close()
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		Expect(got).To(Equal(contract.PositionBottom))
		Expect(string(buf[:n])).To(ContainSubstring("sideways"))
		Expect(string(buf[:n])).To(ContainSubstring("defaulting"))
	})
})

var _ = Describe("loadHighwayConfig (FlagsRowPosition plumbing)", func() {
	const configHome = "root/config"

	loadFromLoader := func(yaml string) *bedrock.FullViewConfig {
		fS := luna.NewMemFS()
		fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
			Data: []byte(yaml),
		}
		loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
		cfg := &bedrock.FullViewConfig{}
		_ = loader.Load("highway", &cfg.Highway)
		return cfg
	}

	Context("when flags-row-position is set in jay.ui.yml", func() {
		It("propagates the value to the ui.HighwayConfig", func() {
			cfg := loadFromLoader(`
ui:
  highway:
    flags-row-position: top
`)
			hCfg := loadHighwayConfig(cfg, contract.SystemPalette())
			Expect(hCfg.FlagsRowPosition).To(Equal(contract.PositionTop))
		})

		It("accepts the bottom value", func() {
			cfg := loadFromLoader(`
ui:
  highway:
    flags-row-position: bottom
`)
			hCfg := loadHighwayConfig(cfg, contract.SystemPalette())
			Expect(hCfg.FlagsRowPosition).To(Equal(contract.PositionBottom))
		})

		It("falls back to bottom for an unrecognised value", func() {
			cfg := loadFromLoader(`
ui:
  highway:
    flags-row-position: sideways
`)
			hCfg := loadHighwayConfig(cfg, contract.SystemPalette())
			Expect(hCfg.FlagsRowPosition).To(Equal(contract.PositionBottom))
		})
	})

	Context("when flags-row-position is absent", func() {
		It("defaults to bottom", func() {
			cfg := loadFromLoader(`
ui:
  highway:
    worker-emoji-pool: '😎'
`)
			hCfg := loadHighwayConfig(cfg, contract.SystemPalette())
			Expect(hCfg.FlagsRowPosition).To(Equal(contract.PositionBottom))
		})
	})
})

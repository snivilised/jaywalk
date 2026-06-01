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
		Expect(normaliseFlagsRowPosition("")).To(Equal(FlagsRowPositionBottom))
	})

	It("returns top unchanged", func() {
		Expect(normaliseFlagsRowPosition(FlagsRowPositionTop)).To(Equal(FlagsRowPositionTop))
	})

	It("returns bottom unchanged", func() {
		Expect(normaliseFlagsRowPosition(FlagsRowPositionBottom)).To(Equal(FlagsRowPositionBottom))
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
		Expect(got).To(Equal(FlagsRowPositionBottom))
		Expect(string(buf[:n])).To(ContainSubstring("sideways"))
		Expect(string(buf[:n])).To(ContainSubstring("defaulting"))
	})
})

var _ = Describe("loadHighwayConfig (FlagsRowPosition plumbing)", func() {
	const configHome = "root/config"

	Context("when flags-row-position is set in jay.ui.yml", func() {
		It("propagates the value to the ui.HighwayConfig", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    flags-row-position: top
`),
			}
			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			cfg, err := loadHighwayConfig(loader, contract.SystemPalette())
			Expect(err).NotTo(HaveOccurred())
			hCfg, ok := cfg.(HighwayConfig)
			Expect(ok).To(BeTrue())
			Expect(hCfg.FlagsRowPosition).To(Equal(FlagsRowPositionTop))
		})

		It("accepts the bottom value", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    flags-row-position: bottom
`),
			}
			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			cfg, err := loadHighwayConfig(loader, contract.SystemPalette())
			Expect(err).NotTo(HaveOccurred())
			hCfg, ok := cfg.(HighwayConfig)
			Expect(ok).To(BeTrue())
			Expect(hCfg.FlagsRowPosition).To(Equal(FlagsRowPositionBottom))
		})

		It("falls back to bottom for an unrecognised value", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    flags-row-position: sideways
`),
			}
			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			cfg, err := loadHighwayConfig(loader, contract.SystemPalette())
			Expect(err).NotTo(HaveOccurred())
			hCfg, ok := cfg.(HighwayConfig)
			Expect(ok).To(BeTrue())
			Expect(hCfg.FlagsRowPosition).To(Equal(FlagsRowPositionBottom))
		})
	})

	Context("when flags-row-position is absent", func() {
		It("defaults to bottom", func() {
			fS := luna.NewMemFS()
			fS.MapFS[configHome+"/jay.ui.yaml"] = &fstest.MapFile{
				Data: []byte(`
ui:
  highway:
    worker-emoji-pool: '😎'
`),
			}
			loader := bedrock.NewViewConfigLoaderWithFS(configHome, fS)
			cfg, err := loadHighwayConfig(loader, contract.SystemPalette())
			Expect(err).NotTo(HaveOccurred())
			hCfg, ok := cfg.(HighwayConfig)
			Expect(ok).To(BeTrue())
			Expect(hCfg.FlagsRowPosition).To(Equal(FlagsRowPositionBottom))
		})
	})
})

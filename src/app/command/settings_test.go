package command

import (
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/mamba/assist"
	"github.com/snivilised/mamba/store"
	"github.com/snivilised/li18ngo"

	"github.com/snivilised/jaywalk/src/locale"
)

// buildTestNavFamilies constructs a fully populated navState bound to a
// real cobra command. Cascade/sampling/poly are bound; nav and preview
// are stubbed nil since createTraversalSettingsIntent only reads from
// cascade, sampling and poly.
func buildTestNavFamilies() (*cobra.Command, *navState) {
	cmd := &cobra.Command{Use: "test"}
	_ = assist.NewParamSet[NavParameterSet](cmd)
	_ = assist.NewParamSet[store.PreviewParameterSet](cmd)
	cascadePs := assist.NewParamSet[store.CascadeParameterSet](cmd)
	cascadePs.Native.BindAll(cascadePs, cmd.Flags())
	samplingPs := assist.NewParamSet[store.SamplingParameterSet](cmd)
	samplingPs.Native.BindAll(samplingPs, cmd.Flags())
	polyPs := assist.NewParamSet[store.PolyFilterParameterSet](cmd)
	polyPs.Native.BindAll(polyPs, cmd.Flags())
	return cmd, &navState{
		cascadeFam:  cascadePs,
		samplingFam: samplingPs,
		polyFam:     polyPs,
	}
}

func init() {
	// mamba's BindAll invokes li18ngo.Text(...) for each flag's
	// usage string; without registration the call panics. We only
	// need the bare Register() here because the test never asserts
	// on rendered flag text.
	_ = li18ngo.Register(
		func(o *li18ngo.UseOptions) {
			o.From.Sources = li18ngo.TranslationFiles{
				locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
			}
		},
	)
}

var _ = Describe("createTraversalSettingsIntent (poly filter Changed detection)", func() {
	It("leaves filter fields empty when no poly flags are set", func() {
		cmd, ns := buildTestNavFamilies()

		cmd.SetArgs([]string{})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.FilesExGlob).To(BeEmpty())
		Expect(intent.Filter.FilesRegEx).To(BeEmpty())
		Expect(intent.Filter.DirectoriesGlob).To(BeEmpty())
		Expect(intent.Filter.DirectoriesRegEx).To(BeEmpty())
	})

	It("captures the --files value when set", func() {
		cmd, ns := buildTestNavFamilies()

		cmd.SetArgs([]string{"--files", "*|.go"})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.FilesExGlob).To(Equal("*|.go"))
		Expect(intent.Filter.FilesRegEx).To(BeEmpty())
		Expect(intent.Filter.DirectoriesGlob).To(BeEmpty())
		Expect(intent.Filter.DirectoriesRegEx).To(BeEmpty())
	})

	It("captures --files-regex when set", func() {
		cmd, ns := buildTestNavFamilies()

		cmd.SetArgs([]string{"--files-regex", `^_`})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.FilesRegEx).To(Equal("^_"))
		Expect(intent.Filter.FilesExGlob).To(BeEmpty())
	})

	It("captures --dirs-glob when set", func() {
		cmd, ns := buildTestNavFamilies()

		cmd.SetArgs([]string{"--dirs-glob", "vendor"})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.DirectoriesGlob).To(Equal("vendor"))
		Expect(intent.Filter.DirectoriesRegEx).To(BeEmpty())
	})

	It("captures --dirs-regex when set", func() {
		cmd, ns := buildTestNavFamilies()

		cmd.SetArgs([]string{"--dirs-regex", `^\.`})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.DirectoriesRegEx).To(Equal(`^\.`))
		Expect(intent.Filter.DirectoriesGlob).To(BeEmpty())
	})

	It("captures multiple flags when several are set", func() {
		cmd, ns := buildTestNavFamilies()

		cmd.SetArgs([]string{
			"--files", "*|.go",
			"--dirs-regex", `^\.`,
		})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.FilesExGlob).To(Equal("*|.go"))
		Expect(intent.Filter.DirectoriesRegEx).To(Equal(`^\.`))
		Expect(intent.Filter.FilesRegEx).To(BeEmpty())
		Expect(intent.Filter.DirectoriesGlob).To(BeEmpty())
	})

	// regression: even if a flag has a non-empty default value, an
	// un-set flag must yield an empty FilterIntent field, so the
	// BenignNodeFilterDef placeholder is correctly distinguished
	// from a user-supplied filter.
	It("treats an un-set flag with a non-empty default as empty in the intent", func() {
		cmd, ns := buildTestNavFamilies()

		// Manually override the default value of the "files" flag to
		// a non-empty string, simulating a future where this flag
		// carries a default. The user's command line did NOT set it.
		flag := cmd.Flags().Lookup(polyFlagFiles)
		Expect(flag).NotTo(BeNil())
		Expect(flag.Value.Set("non-empty-default")).To(Succeed())
		// Important: do NOT mark it as changed.

		cmd.SetArgs([]string{})
		_, err := cmd.ExecuteC()
		Expect(err).NotTo(HaveOccurred())

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.Filter.FilesExGlob).To(BeEmpty(),
			"un-set flag with non-empty default must NOT populate the intent")
	})
})

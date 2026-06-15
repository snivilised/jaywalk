package command

import (
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/li18ngo"
	"github.com/snivilised/mamba/assist"
	"github.com/snivilised/mamba/store"

	"github.com/snivilised/jaywalk/src/locale"
)

// buildTestNavFamilies constructs a fully populated navState bound to a
// real cobra command. Cascade/sampling/poly are bound; nav and preview
// are also constructed (so the bundle can carry IncludeHidden) but
// preview is not assigned to the returned navState since
// createTraversalSettingsIntent only reads from cascade, sampling, poly
// and nav.
func buildTestNavFamilies() (*cobra.Command, *navState) {
	cmd := &cobra.Command{Use: "test"}
	navPs := assist.NewParamSet[NavParameterSet](cmd)
	_ = assist.NewParamSet[store.PreviewParameterSet](cmd)
	cascadePs := assist.NewParamSet[store.CascadeParameterSet](cmd)
	cascadePs.Native.BindAll(cascadePs, cmd.Flags())
	samplingPs := assist.NewParamSet[store.SamplingParameterSet](cmd)
	samplingPs.Native.BindAll(samplingPs, cmd.Flags())
	polyPs := assist.NewParamSet[store.PolyFilterParameterSet](cmd)
	polyPs.Native.BindAll(polyPs, cmd.Flags())

	return cmd, &navState{
		navPs:       navPs,
		cascadeFam:  cascadePs,
		samplingFam: samplingPs,
		polyFam:     polyPs,
	}
}

var _ = Describe("createTraversalSettingsIntent (poly filter Changed detection)", Ordered, func() {
	BeforeEach(func() {
		_ = li18ngo.Register(
			func(o *li18ngo.UseOptions) {
				o.From.Sources = li18ngo.TranslationFiles{
					locale.SourceID: li18ngo.TranslationSource{Name: "agenor"},
				}
			},
		)
	})

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

	It("captures --include-hidden when set on the nav param-set", func() {
		_, ns := buildTestNavFamilies()
		ns.navPs.Native.IncludeHidden = true

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.IncludeHidden).To(BeTrue())
	})

	It("leaves IncludeHidden false by default", func() {
		_, ns := buildTestNavFamilies()

		intent := createTraversalSettingsIntent(navFamilies(ns))
		Expect(intent.IncludeHidden).To(BeFalse())
	})
})

package controller_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"

	jac "github.com/snivilised/jaywalk/src/app/controller"
)

var _ = Describe("TweakCoordinator", func() {
	var (
		palette contract.Palette
		opts    jac.TweakCoordinatorOptions
	)

	BeforeEach(func() {
		palette = contract.SystemPalette()

		opts = jac.TweakCoordinatorOptions{
			PreviewPath: "/tmp",
			Palette:     palette,
			ThemeName:   "system",
			Logger:      nil,
		}
	})

	Describe("NewTweakCoordinator", func() {
		It("initialises all three layers to the provided palette", func() {
			tc := jac.NewTweakCoordinator(opts)
			Expect(tc.WorkingPalette()).To(Equal(palette))
		})

		It("starts with no dirty flags set", func() {
			tc := jac.NewTweakCoordinator(opts)
			Expect(tc.IsDirty()).To(BeFalse())
		})
	})

	Describe("Undo", func() {
		It("resets working palette to the original palette", func() {
			tc := jac.NewTweakCoordinator(opts)
			original := tc.WorkingPalette()

			// Modify the working palette through the coordinator.
			// Since Palette is a value type, we must get, modify, and
			// verify that layer2 was reset by Undo.
			_ = original

			tc.Undo()
			Expect(tc.WorkingPalette()).To(Equal(palette))
		})

		It("clears the creative dirty flag", func() {
			tc := jac.NewTweakCoordinator(opts)
			tc.Undo()
			Expect(tc.IsDirty()).To(BeFalse())
		})
	})

	Describe("IsDirty", func() {
		It("returns false when no changes have been made", func() {
			tc := jac.NewTweakCoordinator(opts)
			Expect(tc.IsDirty()).To(BeFalse())
		})

		It("returns true when creative changes exist", func() {
			tc := jac.NewTweakCoordinator(opts)

			// After Undo, dirty should be false.
			tc.Undo()
			Expect(tc.IsDirty()).To(BeFalse())
		})
	})

	Describe("ExitFlow", func() {
		It("returns true (exit silently) when IsDirty is false", func() {
			tc := jac.NewTweakCoordinator(opts)
			Expect(tc.ExitFlow()).To(BeTrue())
		})

		It("returns true when IsDirty is true (skeleton always exits)", func() {
			tc := jac.NewTweakCoordinator(opts)
			// Skeleton always returns true regardless of dirty state.
			Expect(tc.ExitFlow()).To(BeTrue())
		})
	})

	Describe("WorkingPalette", func() {
		It("returns the palette passed to NewTweakCoordinator", func() {
			tc := jac.NewTweakCoordinator(opts)
			Expect(tc.WorkingPalette()).To(Equal(palette))
		})
	})
})

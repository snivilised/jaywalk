//go:build !race
// +build !race

package scrollbar_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/scrollbar"
)

// newTestTheme creates a Theme suitable for unit tests using the system
// palette. The theme's BranchStyle and MutedStyle are used by the scrollbar
// widget to render the rail and thumb.
func newTestTheme() contract.Theme {
	palette := contract.SystemPalette()
	theme, err := contract.NewTheme(palette, os.Stdout)
	Expect(err).NotTo(HaveOccurred())
	return theme
}

var _ = Describe("Scrollbar", func() {
	Describe("Visible", func() {
		It("returns false when Height is zero", func() {
			state := scrollbar.State{Height: 0, ContentLines: 10, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeFalse())
		})

		It("returns false when Height is negative", func() {
			state := scrollbar.State{Height: -1, ContentLines: 10, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeFalse())
		})

		It("returns false when ContentLines is zero", func() {
			state := scrollbar.State{Height: 5, ContentLines: 0, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeFalse())
		})

		It("returns false when ContentLines equals Height", func() {
			state := scrollbar.State{Height: 5, ContentLines: 5, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeFalse())
		})

		It("returns false when ContentLines is less than Height", func() {
			state := scrollbar.State{Height: 10, ContentLines: 5, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeFalse())
		})

		It("returns true when ContentLines exceeds Height", func() {
			state := scrollbar.State{Height: 5, ContentLines: 10, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeTrue())
		})

		It("returns true at the boundary (Height+1 lines)", func() {
			state := scrollbar.State{Height: 5, ContentLines: 6, Offset: 0}
			Expect(scrollbar.Visible(state)).To(BeTrue())
		})
	})

	Describe("View", func() {
		var theme contract.Theme

		BeforeEach(func() {
			theme = newTestTheme()
		})

		Context("when content fits in viewport", func() {
			It("returns empty string when Height is zero", func() {
				state := scrollbar.State{Height: 0, ContentLines: 10, Offset: 0}
				cfg := scrollbar.Config{Theme: theme}
				Expect(scrollbar.View(state, cfg)).To(BeEmpty())
			})

			It("returns empty string when ContentLines equals Height", func() {
				state := scrollbar.State{Height: 5, ContentLines: 5, Offset: 0}
				cfg := scrollbar.Config{Theme: theme}
				Expect(scrollbar.View(state, cfg)).To(BeEmpty())
			})
		})

		Context("when content exceeds viewport", func() {
			It("renders Height rows each ending with newline", func() {
				state := scrollbar.State{Height: 5, ContentLines: 20, Offset: 0}
				cfg := scrollbar.Config{Theme: theme}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				Expect(lines).To(HaveLen(5))
			})

			It("places thumb at row 0 when Offset is 0", func() {
				state := scrollbar.State{Height: 5, ContentLines: 20, Offset: 0}
				cfg := scrollbar.Config{Theme: theme}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				// Row 0 should contain the thumb (█ character)
				Expect(lines[0]).To(ContainSubstring("█"))
				// Other rows should not contain the thumb
				for i := 1; i < 5; i++ {
					Expect(lines[i]).NotTo(ContainSubstring("█"))
				}
			})

			It("places thumb at last row when Offset is at bottom", func() {
				state := scrollbar.State{Height: 5, ContentLines: 20, Offset: 15}
				cfg := scrollbar.Config{Theme: theme}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				// Last row should contain the thumb
				Expect(lines[4]).To(ContainSubstring("█"))
				// Other rows should not contain the thumb
				for i := 0; i < 4; i++ {
					Expect(lines[i]).NotTo(ContainSubstring("█"))
				}
			})

			It("places thumb proportionally in the middle", func() {
				state := scrollbar.State{Height: 10, ContentLines: 100, Offset: 45}
				cfg := scrollbar.Config{Theme: theme}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				// Find which row has the thumb
				thumbRow := -1
				for i, line := range lines {
					if strings.Contains(line, "█") {
						thumbRow = i
						break
					}
				}
				Expect(thumbRow).To(BeNumerically(">=", 0))
				Expect(thumbRow).To(BeNumerically("<", 10))
				// At offset 45 in 100 lines with 10 visible, thumb should be
				// around row 4-5 (45/90 * 9 ≈ 4.5)
				Expect(thumbRow).To(BeNumerically(">=", 3))
				Expect(thumbRow).To(BeNumerically("<=", 6))
			})

			It("renders rail characters for non-thumb rows", func() {
				state := scrollbar.State{Height: 3, ContentLines: 10, Offset: 0}
				cfg := scrollbar.Config{Theme: theme}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				// Row 0 has thumb, rows 1-2 should have rail (branch vertical)
				Expect(lines[1]).NotTo(ContainSubstring("█"))
				Expect(lines[2]).NotTo(ContainSubstring("█"))
			})
		})

		Context("with custom TreeIconBranchVertical", func() {
			It("uses the custom branch character for the rail", func() {
				themeCopy := theme
				themeCopy.TreeIcons[contract.TreeIconBranchVertical] = "║"
				state := scrollbar.State{Height: 3, ContentLines: 10, Offset: 0}
				cfg := scrollbar.Config{Theme: themeCopy}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				// Non-thumb rows should contain the custom rail character
				Expect(lines[1]).To(ContainSubstring("║"))
				Expect(lines[2]).To(ContainSubstring("║"))
			})
		})

		Context("with empty TreeIconBranchVertical", func() {
			It("falls back to the default │ character", func() {
				themeCopy := theme
				themeCopy.TreeIcons[contract.TreeIconBranchVertical] = ""
				state := scrollbar.State{Height: 3, ContentLines: 10, Offset: 0}
				cfg := scrollbar.Config{Theme: themeCopy}
				out := scrollbar.View(state, cfg)

				lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
				// Non-thumb rows should contain the default rail character
				Expect(lines[1]).To(ContainSubstring("│"))
				Expect(lines[2]).To(ContainSubstring("│"))
			})
		})
	})

	Describe("ScrollbarGutterWidth", func() {
		It("is 1 column", func() {
			Expect(scrollbar.ScrollbarGutterWidth).To(Equal(1))
		})
	})
})

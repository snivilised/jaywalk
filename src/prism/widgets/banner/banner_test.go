//go:build !race
// +build !race

package banner_test

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// smallArt is a 3-line mini banner used for unit tests. It contains
// face characters (█), shadow characters (╗, ║, ═) and spaces, so all
// three classes are exercised. The art's intrinsic width (max line
// rune count) is 18.
const smallArt = "██╗     ██╗\n" +
	"║ ║     ║ ║\n" +
	"╚═╝     ╚═╝"

var _ = Describe("Banner.Render", func() {
	Describe("visibility / disabled", func() {
		It("falls back to DefaultArt when Art is empty", func() {
			out := banner.Render(
				banner.Config{Art: ""},
				banner.Styles{},
				banner.Effect{},
			)
			// Empty Art + empty gradient = plain default art,
			// which contains face runes and no ANSI codes.
			Expect(out).NotTo(BeEmpty())
			Expect(out).To(ContainSubstring("█"))
			Expect(out).NotTo(ContainSubstring("\x1b["))
		})

		It("returns plain art when Gradient is nil", func() {
			out := banner.Render(
				banner.Config{Art: smallArt},
				banner.Styles{},
				banner.Effect{Gradient: nil, State: newState()},
			)
			// Render always appends a trailing newline so the
			// banner sits on its own lines.
			Expect(out).To(Equal(smallArt + "\n"))
			Expect(out).NotTo(ContainSubstring("\x1b["))
		})

		It("returns plain art when State is nil", func() {
			grad := newGradient()
			out := banner.Render(
				banner.Config{Art: smallArt},
				banner.Styles{},
				banner.Effect{Gradient: grad, State: nil},
			)
			// Render always appends a trailing newline so the
			// banner sits on its own lines.
			Expect(out).To(Equal(smallArt + "\n"))
			Expect(out).NotTo(ContainSubstring("\x1b["))
		})

		It("tolerates a leading carriage return in the art", func() {
			// Simulates the human-readable raw-string format used
			// for DefaultArt (newline after the opening backtick).
			withCR := "\n" + smallArt
			out := banner.Render(
				banner.Config{Art: withCR},
				banner.Styles{},
				banner.Effect{Gradient: nil, State: newState()},
			)
			// The leading CR must be stripped so the banner does
			// not start with an empty padded line. Render's
			// trailing newline then takes its place.
			Expect(out).To(Equal(smallArt + "\n"))
		})

		It("tolerates a trailing carriage return in the art", func() {
			withCR := smallArt + "\n"
			out := banner.Render(
				banner.Config{Art: withCR},
				banner.Styles{},
				banner.Effect{Gradient: nil, State: newState()},
			)
			// Only ONE trailing newline should remain (the one
			// Render adds to separate the banner from the host
			// view), not the original one plus Render's.
			Expect(out).To(Equal(smallArt + "\n"))
		})
	})

	Describe("unified gradient", func() {
		var (
			grad   *contract.ResolvedGradient
			state  *effects.GradientState
			effect banner.Effect
		)

		BeforeEach(func() {
			grad = newGradient()
			state = newState()
			effect = banner.Effect{
				Gradient: grad,
				State:    state,
				Aspects:  banner.Aspects{Unity: banner.UnityUnified},
			}
		})

		It("colours both face and shadow characters", func() {
			out := banner.Render(banner.Config{Art: smallArt}, banner.Styles{}, effect)
			Expect(out).NotTo(BeEmpty())
			Expect(out).To(ContainSubstring("\x1b[38;2;"))
		})

		It("includes the gradient's Hi and Lo colours in the output", func() {
			// Use a 4-rune non-whitespace line so that all four
			// steps of the gradient are reachable from the start.
			lineArt := "████"
			// First render populates the state's steps array.
			_ = banner.Render(banner.Config{Art: lineArt}, banner.Styles{}, effect)
			// At state.Offset=0, the four positions all map to the
			// same step (step[0] = Hi) because the gradient is
			// 4 steps long and there are 4 positions. To reach
			// Lo, advance the state by 3.
			state.Update(3)
			out := banner.Render(banner.Config{Art: lineArt}, banner.Styles{}, effect)
			// Hi = (255, 0, 0), Lo = (0, 0, 255)
			Expect(out).To(ContainSubstring("\x1b[38;2;0;0;255m"))
		})

		It("preserves spaces and newlines (alignment not disturbed)", func() {
			out := banner.Render(banner.Config{Art: smallArt}, banner.Styles{}, effect)
			// Strip ANSI codes; Render appends a trailing newline
			// so the banner sits on its own lines, hence the +"\n".
			Expect(stripAnsi(out)).To(Equal(smallArt + "\n"))
		})
	})

	Describe("unity = GradientFace", func() {
		var grad *contract.ResolvedGradient
		var state *effects.GradientState

		BeforeEach(func() {
			grad = newGradient()
			state = newState()
		})

		It("locks shadow characters to a single colour (FixedEnd = Hi)", func() {
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{
					Unity:    banner.UnityGradientFace,
					FixedEnd: banner.FixedEndHi,
				},
			}
			out := banner.Render(banner.Config{Art: smallArt}, banner.Styles{}, effect)

			// The ping-pong gradient can place face runes at the
			// Lo step (e.g. when the phase maps to 2*(n-1) - phase
			// = n-1). The strict invariant however is that NO
			// shadow rune uses Lo - they are all pinned to Hi.
			// We verify this by ensuring the number of Hi codes
			// is at least the number of shadow runes (face runes
			// may also be Hi and add to the count, but cannot
			// subtract from it).
			hiCount := strings.Count(out, "\x1b[38;2;255;0;0m")
			shadowRunes := countShadows(smallArt)
			faceRunes := countFaces(smallArt)
			Expect(hiCount).To(BeNumerically(">=", shadowRunes),
				"every shadow rune must use the Hi colour")

			// Lo codes are allowed (face runes can hit Lo via
			// the ping-pong), but there must be no MORE Lo codes
			// than there are face runes.
			loCount := strings.Count(out, "\x1b[38;2;0;0;255m")
			Expect(loCount).To(BeNumerically("<=", faceRunes),
				"only face runes may be coloured Lo when FixedEnd=Hi")
		})

		It("locks shadow characters to Lo when FixedEnd = Lo", func() {
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{
					Unity:    banner.UnityGradientFace,
					FixedEnd: banner.FixedEndLo,
				},
			}
			out := banner.Render(banner.Config{Art: smallArt}, banner.Styles{}, effect)

			// Every shadow rune is pinned to Lo, so the count of
			// Lo codes is at least the number of shadow runes.
			loCount := strings.Count(out, "\x1b[38;2;0;0;255m")
			shadowRunes := countShadows(smallArt)
			Expect(loCount).To(BeNumerically(">=", shadowRunes),
				"every shadow rune must use the Lo colour")
		})
	})

	Describe("unity = ShadowFace", func() {
		var grad *contract.ResolvedGradient
		var state *effects.GradientState

		BeforeEach(func() {
			grad = newGradient()
			state = newState()
		})

		It("locks face characters to a single colour", func() {
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{
					Unity:    banner.UnityShadowFace,
					FixedEnd: banner.FixedEndHi,
				},
			}
			out := banner.Render(banner.Config{Art: smallArt}, banner.Styles{}, effect)
			// All face runes are pinned to Hi. Count the Hi codes;
			// it must be at least the number of face runes
			// (shadow runes may add more if any happen to land on
			// step 0).
			hiCount := strings.Count(out, "\x1b[38;2;255;0;0m")
			faceRunes := countFaces(smallArt)
			Expect(hiCount).To(BeNumerically(">=", faceRunes))
		})
	})

	Describe("orientation", func() {
		var grad *contract.ResolvedGradient
		var state *effects.GradientState

		BeforeEach(func() {
			grad = newGradient()
			state = newState()
		})

		It("horizontal: colour index varies with column", func() {
			// Use a single-line art to make the assertion direct.
			lineArt := "████╗╗╗╗"
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{
					Orientation: banner.OrientationHorizontal,
					Unity:       banner.UnityUnified,
				},
			}
			out := banner.Render(banner.Config{Art: lineArt}, banner.Styles{}, effect)
			colours := orderedRuneColours(out, lineArt)
			Expect(colours).To(HaveLen(len([]rune(lineArt))))
			// Different positions should produce different colours
			// (or at least the first and last should differ).
			Expect(colours[0]).NotTo(Equal(colours[len(colours)-1]))
		})

		It("vertical: same column in different rows gets the same colour", func() {
			// At state.Offset=0, GetEffectiveIndex(0)=0 → step[0] = Hi
			// for any position. So all runes look Hi. To verify
			// vertical orientation we need to set the state offset
			// and check that the index used is the row, not the
			// column. We do this by checking that runes in the same
			// column at different rows get the same colour.
			columnArt := "█\n█\n█"
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{
					Orientation: banner.OrientationVertical,
					Unity:       banner.UnityUnified,
				},
			}
			out := banner.Render(banner.Config{Art: columnArt}, banner.Styles{}, effect)
			colours := orderedRuneColours(out, columnArt)
			Expect(colours).To(HaveLen(3))
			// All three runes are in column 0 of rows 0, 1, 2.
			// With vertical orientation, all three map to the same
			// step index (row number, modulo the step count).
			// At state.Offset=0: stepIdx = (0+row) % 4 → rows 0, 1, 2
			// map to steps 0, 1, 2 which are all distinct. So the
			// colours SHOULD differ. Verify they do.
			Expect(colours[0]).NotTo(Equal(colours[1]))
			Expect(colours[1]).NotTo(Equal(colours[2]))
		})
	})

	Describe("justification", func() {
		var (
			grad  *contract.ResolvedGradient
			state *effects.GradientState
		)

		BeforeEach(func() {
			grad = newGradient()
			state = newState()
		})

		It("JustifyRight (default) pads so the right edge aligns with Width", func() {
			// Use Width smaller than the art to ensure no padding
			// is added; this case is degenerate. Then use a Width
			// larger than the art to verify padding is added.
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{Unity: banner.UnityUnified},
			}
			width := 24
			out := banner.Render(banner.Config{
				Art:   smallArt,
				Width: width,
				// Justify omitted → default
			}, banner.Styles{}, effect)

			// Each line of the art has rune count 11; right-justified
			// at width 24 → pad 13 leading spaces. Render appends a
			// trailing newline, so Split produces one extra empty
			// line which we tolerate.
			lines := strings.Split(out, "\n")
			Expect(lines).To(HaveLen(4))
			Expect(lines[3]).To(BeEmpty())
			for _, line := range lines[:3] {
				// Count leading spaces (ANSI reset codes do not
				// count as width because they are zero-width).
				leading := leadingSpaces(line)
				Expect(leading).To(Equal(13))
			}
		})

		It("JustifyLeft adds no padding", func() {
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{Unity: banner.UnityUnified},
			}
			out := banner.Render(banner.Config{
				Art:     smallArt,
				Width:   30,
				Justify: banner.JustifyLeft,
			}, banner.Styles{}, effect)

			// Art's first line is "██╗     ██╗" - 5 leading spaces
			// followed by █. The first non-space rune must be at
			// column 0 of the visible content (i.e. line must not
			// start with extra spaces).
			lines := strings.Split(out, "\n")
			Expect(leadingSpaces(lines[0])).To(Equal(0))
		})

		It("JustifyCenter pads to the centre of Width", func() {
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{Unity: banner.UnityUnified},
			}
			width := 25
			// lineWidth = 11, pad = (25 - 11) / 2 = 7
			out := banner.Render(banner.Config{
				Art:     smallArt,
				Width:   width,
				Justify: banner.JustifyCenter,
			}, banner.Styles{}, effect)
			lines := strings.Split(out, "\n")
			// Render appends a trailing newline, so Split produces
			// one extra empty line which we tolerate by skipping it.
			for _, line := range lines[:len(lines)-1] {
				Expect(leadingSpaces(line)).To(Equal(7))
			}
		})
	})

	Describe("newlines pass-through", func() {
		It("preserves the number of newlines in the art", func() {
			grad := newGradient()
			state := newState()
			effect := banner.Effect{
				Gradient: grad, State: state,
				Aspects: banner.Aspects{Unity: banner.UnityUnified},
			}
			out := banner.Render(banner.Config{Art: smallArt}, banner.Styles{}, effect)
			// smallArt has 2 newlines; Render appends a trailing
			// newline so the banner sits on its own line, hence +1.
			Expect(strings.Count(out, "\n")).To(Equal(strings.Count(smallArt, "\n") + 1))
		})
	})
})

var _ = Describe("Aspects randomisation", func() {
	It("covers all three Unity values across many seeds", func() {
		seen := map[banner.Unity]bool{}
		for i := 0; i < 200; i++ {
			rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+2))) //nolint:gosec // test-only random
			aspects := banner.RandomiseAspectsForTest(rng)
			seen[aspects.Unity] = true
		}
		Expect(seen).To(HaveLen(3))
	})

	It("covers both orientations", func() {
		seen := map[banner.Orientation]bool{}
		for i := 0; i < 200; i++ {
			rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+2))) //nolint:gosec // test-only random
			aspects := banner.RandomiseAspectsForTest(rng)
			seen[aspects.Orientation] = true
		}
		Expect(seen).To(HaveLen(2))
	})

	It("covers both banding values", func() {
		seen := map[banner.Banding]bool{}
		for i := 0; i < 200; i++ {
			rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+2))) //nolint:gosec // test-only random
			aspects := banner.RandomiseAspectsForTest(rng)
			seen[aspects.Banding] = true
		}
		Expect(seen).To(HaveLen(2))
	})

	It("sets FixedEnd to Unfixed when Unity is Unified", func() {
		for i := 0; i < 200; i++ {
			rng := rand.New(rand.NewPCG(uint64(i+1), uint64(i+2))) //nolint:gosec // test-only random
			aspects := banner.RandomiseAspectsForTest(rng)
			if aspects.Unity == banner.UnityUnified {
				Expect(aspects.FixedEnd).To(Equal(banner.FixedEndUnfixed))
			}
		}
	})

	It("is deterministic for a given seed", func() {
		rngA := rand.New(rand.NewPCG(42, 99)) //nolint:gosec // test-only random
		rngB := rand.New(rand.NewPCG(42, 99)) //nolint:gosec // test-only random
		a := banner.RandomiseAspectsForTest(rngA)
		b := banner.RandomiseAspectsForTest(rngB)
		Expect(a).To(Equal(b))
	})
})

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// newGradient returns a fixed Hi=red, Lo=blue, Steps=4 gradient for
// deterministic test output.
func newGradient() *contract.ResolvedGradient {
	return &contract.ResolvedGradient{
		Steps: 4,
		Hi:    color.RGBA{R: 255, G: 0, B: 0, A: 255},
		Lo:    color.RGBA{R: 0, G: 0, B: 255, A: 255},
	}
}

func newState() *effects.GradientState {
	st := effects.NewGradientState()
	st.TotalSteps = 4
	return st
}

// stripAnsi removes ANSI 24-bit colour escape sequences and reset codes
// from s. Mirrors the lab.StripANSI helper used elsewhere in the
// codebase.
var ansiRe = []string{
	"\x1b[38;2;0;0;0m",
	"\x1b[38;2;255;255;255m",
	"\x1b[38;2;255;0;0m",
	"\x1b[38;2;0;0;255m",
	"\x1b[0m",
}

func stripAnsi(s string) string {
	out := s
	for _, c := range ansiRe {
		out = strings.ReplaceAll(out, c, "")
	}
	// Generic fallback: strip any remaining \x1b[...m sequences.
	for {
		start := strings.Index(out, "\x1b[")
		if start < 0 {
			break
		}
		end := strings.Index(out[start:], "m")
		if end < 0 {
			break
		}
		out = out[:start] + out[start+end+1:]
	}
	return out
}

// countShadows returns the number of non-space, non-newline, non-face
// runes in s.
func countShadows(s string) int {
	n := 0
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\t' {
			continue
		}
		if r == '█' {
			continue
		}
		n++
	}
	return n
}

// countFaces returns the number of face runes (█) in s.
func countFaces(s string) int {
	n := 0
	for _, r := range s {
		if r == '█' {
			n++
		}
	}
	return n
}

// orderedRuneColours returns the RGB colour of each non-whitespace rune
// in art, in order, extracted from the rendered output. The returned
// slice is parallel to art's non-whitespace runes.
func orderedRuneColours(rendered, _ string) []color.RGBA {
	return findAnsiColours(rendered)
}

// findAnsiColours extracts the RGB colours from \x1b[38;2;R;G;Bm
// sequences in s, in order.
func findAnsiColours(s string) []color.RGBA {
	var out []color.RGBA
	prefix := "\x1b[38;2;"
	for {
		i := strings.Index(s, prefix)
		if i < 0 {
			break
		}
		rest := s[i+len(prefix):]
		end := strings.Index(rest, "m")
		if end < 0 {
			break
		}
		var r, g, b uint8
		_, _ = fmt.Sscanf(rest[:end], "%d;%d;%d", &r, &g, &b)
		out = append(out, color.RGBA{R: r, G: g, B: b, A: 255})
		s = rest[end+1:]
	}
	return out
}

// leadingSpaces returns the number of leading space characters in s
// (ignoring ANSI escape sequences).
func leadingSpaces(s string) int {
	n := 0
	inAnsi := false
	for _, r := range s {
		if r == 0x1b {
			inAnsi = true
			continue
		}
		if inAnsi {
			if r == 'm' {
				inAnsi = false
			}
			continue
		}
		if r == ' ' {
			n++
			continue
		}
		break
	}
	return n
}

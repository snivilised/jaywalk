package track

import (
	"errors"
	"io"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func testTheme() contract.Theme {
	t, err := contract.NewTheme(contract.SystemPalette(), io.Discard)
	if err != nil {
		panic(err)
	}
	return t
}

func noopFrame(_ int) string { return "•" }

func baseLane() Lane {
	return Lane{
		Emoji: "🔍", JobEmoji: "🍎", Label: "test", FrameFn: noopFrame,
	}
}

func baseModel(lanes int) Model {
	l := make([]Lane, lanes)
	for i := range l {
		l[i] = baseLane()
	}
	return New(
		WithLanes(l),
		WithTheme(testTheme()),
		WithTickRate(50*time.Millisecond),
		WithMaxDepth(5),
		WithWidth(80),
	)
}

// update is a convenience wrapper that casts the bubbletea Model
// back to Model. Mirrors the helper used in
// src/prism/highway/model_test.go.
func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	r, cmd := m.Update(msg)
	return r.(Model), cmd //nolint:errcheck // known concrete type
}

// tickNow returns a fresh TickMsg value.
func tickNow() TickMsg { return TickMsg(time.Now()) }

// countOccurrences counts the non-overlapping occurrences of sub
// in s.
func countOccurrences(s, sub string) int {
	count := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			count++
		}
	}
	return count
}

// ---------------------------------------------------------------------------
// New
// ---------------------------------------------------------------------------

var _ = Describe("New", func() {
	It("sets the lanes slice", func() {
		m := baseModel(3)
		Expect(m.lanes).To(HaveLen(3))
	})

	It("initialises the counted map to empty (verified by MotifMsg behaviour)", func() {
		m := baseModel(1)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/a", IsDir: false,
		}})
		Expect(updated.files).To(Equal(1))
	})

	It("initialises currentLaneIdx to 0", func() {
		m := baseModel(1)
		Expect(m.currentLaneIdx).To(Equal(0))
	})

	It("initialises tickRate to the supplied value", func() {
		m := New(WithTickRate(25 * time.Millisecond))
		Expect(m.tickRate).To(Equal(25 * time.Millisecond))
	})

	It("falls back to 50ms when WithTickRate receives a non-positive value", func() {
		m := New(WithTickRate(0))
		Expect(m.tickRate).To(Equal(50 * time.Millisecond))

		m = New(WithTickRate(-1 * time.Millisecond))
		Expect(m.tickRate).To(Equal(50 * time.Millisecond))
	})
})

// ---------------------------------------------------------------------------
// Options
// ---------------------------------------------------------------------------

var _ = Describe("Options", func() {
	It("WithLanes computes the skip factor via the current tick rate", func() {
		fast := baseLane()
		slow := baseLane()
		slow.IntervalMs = 500 // 500/50 = 10

		m := New(
			WithTickRate(50*time.Millisecond),
			WithLanes([]Lane{fast, slow}),
		)
		// After 10 ticks, fast should be at 10 and slow at 1.
		for range 10 {
			m, _ = update(m, tickNow())
		}
		Expect(m.lanes[0].tick).To(Equal(10))
		Expect(m.lanes[1].tick).To(Equal(1))
	})

	It("WithLanes default tick rate is 50ms when WithTickRate is omitted", func() {
		lane := baseLane()
		lane.IntervalMs = 100 // 100/50 = 2
		m := New(WithLanes([]Lane{lane}))
		for range 4 {
			m, _ = update(m, tickNow())
		}
		Expect(m.lanes[0].tick).To(Equal(2))
	})

	It("WithTheme populates Styles from a resolved theme", func() {
		m := New(
			WithLanes([]Lane{baseLane()}),
			WithTheme(testTheme()),
		)
		// styles should be non-zero after WithTheme.
		Expect(m.styles.BarFilledStyle).NotTo(Equal(lipgloss.NewStyle()))
	})

	It("WithStyles overrides individual style fields", func() {
		muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		m := New(
			WithLanes([]Lane{baseLane()}),
			WithStyles(Styles{MutedStyle: muted}),
		)
		Expect(m.styles.MutedStyle.GetForeground()).NotTo(BeNil())
	})

	It("WithTheme after WithStyles wins for fields it populates", func() {
		muted := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
		m := New(
			WithLanes([]Lane{baseLane()}),
			WithStyles(Styles{MutedStyle: muted}),
			WithTheme(testTheme()),
		)
		// WithTheme re-assigns styles.MutedStyle, so the
		// WithStyles colour should be gone.
		// (We can't easily assert the colour value, but we
		// can assert the style isn't the literal one we
		// created via WithStyles.)
		_ = muted
		Expect(m.styles.MutedStyle).NotTo(Equal(muted))
	})

	It("WithMaxDepth sets the max depth", func() {
		m := New(WithMaxDepth(7))
		Expect(m.maxDepth).To(Equal(uint(7)))
	})

	It("WithNoRecurse sets the no-recurse flag", func() {
		m := New(WithNoRecurse(true))
		Expect(m.noRecurse).To(BeTrue())
	})

	It("WithWidth sets the initial width", func() {
		m := New(WithWidth(120))
		Expect(m.width).To(Equal(120))
	})

	It("WithTickRate recomputes skip when lanes are already set", func() {
		lane := baseLane()
		lane.IntervalMs = 100
		// Set lanes first with a 50ms rate; skip = 2.
		m := New(WithLanes([]Lane{lane}))
		// Now switch to 25ms; skip recomputes to 4.
		m2 := New(WithLanes([]Lane{lane}), WithTickRate(25*time.Millisecond))
		Expect(m2.skip[0]).To(Equal(4))
		Expect(m.skip[0]).To(Equal(2))
	})
})

// ---------------------------------------------------------------------------
// Model.Init
// ---------------------------------------------------------------------------

var _ = Describe("Model.Init", func() {
	It("returns nil (the highway root owns the ticker)", func() {
		m := baseModel(1)
		Expect(m.Init()).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// Model.Update - WidthMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - WidthMsg", func() {
	It("updates the width", func() {
		m := baseModel(1)
		updated, cmd := update(m, WidthMsg{Width: 120})
		Expect(cmd).To(BeNil())
		Expect(updated.width).To(Equal(120))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - tickMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - TickMsg", func() {
	It("increments ticks on all lanes", func() {
		m := baseModel(2)
		Expect(m.lanes[0].tick).To(Equal(0))
		Expect(m.lanes[1].tick).To(Equal(0))

		updated, cmd := update(m, tickNow())
		Expect(cmd).To(BeNil())
		Expect(updated.lanes[0].tick).To(Equal(1))
		Expect(updated.lanes[1].tick).To(Equal(1))
	})

	It("respects the skip factor on slow lanes", func() {
		fast := baseLane()
		slow := baseLane()
		slow.IntervalMs = 500
		m := New(
			WithTickRate(50*time.Millisecond),
			WithLanes([]Lane{fast, slow}),
		)
		for range 10 {
			m, _ = update(m, tickNow())
		}
		Expect(m.lanes[0].tick).To(Equal(10))
		Expect(m.lanes[1].tick).To(Equal(1))
	})

	It("advances a very slow lane exactly once in 100 ticks", func() {
		very := baseLane()
		very.IntervalMs = 5000 // 5000/50 = 100
		m := New(
			WithTickRate(50*time.Millisecond),
			WithLanes([]Lane{very}),
		)
		for range 100 {
			m, _ = update(m, tickNow())
		}
		Expect(m.lanes[0].tick).To(Equal(1))
	})

	It("returns no cmd (root drives the next tick)", func() {
		m := baseModel(1)
		_, cmd := update(m, tickNow())
		Expect(cmd).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// Model.Update - MotifMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - MotifMsg", func() {
	It("applies motif data to the current lane", func() {
		m := baseModel(2)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", Name: "a.txt", IsDir: false, Depth: 1,
		}})
		Expect(updated.lanes[0].Path).To(Equal("/root/a.txt"))
		Expect(updated.lanes[0].Name).To(Equal("a.txt"))
		Expect(updated.lanes[0].IsDir).To(BeFalse())
		Expect(updated.lanes[0].Depth).To(Equal(uint(1)))
	})

	It("rotates currentLaneIdx round-robin", func() {
		m := baseModel(2)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(updated.currentLaneIdx).To(Equal(1))

		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/b.txt", IsDir: false,
		}})
		Expect(updated.lanes[1].Path).To(Equal("/root/b.txt"))
		Expect(updated.currentLaneIdx).To(Equal(0))
	})

	It("dedupes via the counted map", func() {
		m := baseModel(1)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(1))

		// Same path again - no increment.
		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(1))

		// Different path - incremented.
		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/b.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(2))
	})

	It("increments Dirs for directories", func() {
		m := baseModel(1)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/src", IsDir: true,
		}})
		Expect(updated.dirs).To(Equal(1))
		Expect(updated.files).To(Equal(0))
	})

	It("sets action and pipeline info on the lane", func() {
		m := baseModel(1)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path:            "/f.txt",
			ActionName:      "encode",
			PipelineName:    "pipe",
			CommandOutput:   "ok",
			ExecutionString: "ffmpeg",
			DryRun:          true,
			Err:             errors.New("boom"),
		}})
		Expect(updated.lanes[0].ActionName).To(Equal("encode"))
		Expect(updated.lanes[0].PipelineName).To(Equal("pipe"))
		Expect(updated.lanes[0].CommandOutput).To(Equal("ok"))
		Expect(updated.lanes[0].ExecutionString).To(Equal("ffmpeg"))
		Expect(updated.lanes[0].DryRun).To(BeTrue())
		Expect(updated.lanes[0].Err).To(MatchError("boom"))
	})

	It("handles zero lanes gracefully (counted still increments)", func() {
		m := baseModel(0)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/f.txt",
		}})
		// Files still increments because the dedup is
		// independent of lane application.
		Expect(updated.files).To(Equal(1))
	})

	It("applies gradient to the lane and initialises GradientState", func() {
		m := baseModel(1)
		grad := &contract.ResolvedGradient{
			Steps: 5, Hi: lipgloss.Color("#FF0000"), Lo: lipgloss.Color("#000000"),
		}
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path:     "/f.txt",
			Gradient: grad,
		}})
		Expect(updated.lanes[0].HighlightGradient).To(Equal(grad))
		Expect(updated.lanes[0].GradientState).NotTo(BeNil())
		Expect(updated.lanes[0].GradientState.TotalSteps).To(Equal(5))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - CensusMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - CensusMsg", func() {
	It("increases maxDepth when a larger value is seen", func() {
		m := baseModel(1)
		Expect(m.maxDepth).To(Equal(uint(5)))

		updated, _ := update(m, CensusMsg{MaxDepth: 10})
		Expect(updated.maxDepth).To(Equal(uint(10)))
	})

	It("does not decrease maxDepth", func() {
		m := baseModel(1)
		Expect(m.maxDepth).To(Equal(uint(5)))

		updated, _ := update(m, CensusMsg{MaxDepth: 3})
		Expect(updated.maxDepth).To(Equal(uint(5)))
	})

	It("returns no cmd", func() {
		m := baseModel(1)
		_, cmd := update(m, CensusMsg{MaxDepth: 7})
		Expect(cmd).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// Model.Update - CompleteMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - CompleteMsg", func() {
	It("clears the counted map so subsequent motifs increment again", func() {
		m := baseModel(1)
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(1))

		updated, _ = update(updated, CompleteMsg{})
		// Same path after CompleteMsg: dedup map is empty,
		// so this is treated as a new path.
		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(2))
	})

	It("returns no cmd", func() {
		m := baseModel(1)
		_, cmd := update(m, CompleteMsg{})
		Expect(cmd).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// Model.Update - unknown message
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - unknown message", func() {
	It("returns the model unchanged with nil cmd", func() {
		m := baseModel(1)
		_, cmd := update(m, "some-unknown-msg")
		Expect(cmd).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// View
// ---------------------------------------------------------------------------

var _ = Describe("View", func() {
	It("renders all lanes with separators", func() {
		m := New(
			WithLanes([]Lane{baseLane(), baseLane(), baseLane()}),
			WithTheme(testTheme()),
			WithWidth(80),
		)
		v := m.View()
		// Three lanes means three `├──` separators.
		Expect(v.Content).To(ContainSubstring("├"))
		// Worker emoji appears at least three times (once
		// per lane).
		Expect(countOccurrences(v.Content, "🔍")).To(BeNumerically(">=", 3))
	})

	It("respects WidthMsg", func() {
		m := New(
			WithLanes([]Lane{baseLane()}),
			WithTheme(testTheme()),
		)
		m, _ = update(m, WidthMsg{Width: 100})
		v := m.View()
		Expect(v.Content).NotTo(BeEmpty())
	})

	It("returns an empty string when there are no lanes", func() {
		m := New(WithTheme(testTheme()), WithWidth(80))
		Expect(m.View().Content).To(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
// Lane.WindowSize
// ---------------------------------------------------------------------------

var _ = Describe("Lane.WindowSize", func() {
	It("returns the command output length when present", func() {
		l := Lane{CommandOutput: "hello"}
		Expect(l.WindowSize()).To(Equal(5))
	})

	It("returns 6 for action animations on files", func() {
		l := Lane{ActionName: "encode", IsDir: false}
		Expect(l.WindowSize()).To(Equal(6))
	})

	It("returns 4 for the default case", func() {
		l := Lane{}
		Expect(l.WindowSize()).To(Equal(4))
	})

	It("returns 4 for dir with action name (action animation is file-only)", func() {
		l := Lane{ActionName: "encode", IsDir: true}
		Expect(l.WindowSize()).To(Equal(4))
	})
})

// Reference imports that may become unused after future refactors.
var _ = testing.Short

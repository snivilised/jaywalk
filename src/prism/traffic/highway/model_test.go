package highway

import (
	"errors"
	"io"
	"time"

	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/traffic"
	"github.com/snivilised/jaywalk/src/prism/widget"
)

// ---------------------------------------------------------------------------
// Helpers
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
	return Lane{Emoji: "🔍", Label: "test", FrameFn: noopFrame}
}

func baseModel(lanes int) Model {
	l := make([]Lane, lanes)
	for i := range l {
		l[i] = baseLane()
	}
	return NewModel(l, 50*time.Millisecond, "/root", 5, testTheme(), false)
}

// update is a convenience wrapper that calls Update on the model and
// type-asserts the result back to Model.
func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	r, cmd := m.Update(msg)
	return r.(Model), cmd //nolint:errcheck // ok
}

// invokeCmd executes a tea.Cmd and returns the resulting message.
func invokeCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// ---------------------------------------------------------------------------
// NewModel
// ---------------------------------------------------------------------------

var _ = Describe("NewModel", func() {
	It("sets the lanes slice", func() {
		m := baseModel(3)
		Expect(m.lanes).To(HaveLen(3))
	})

	It("sets the tick rate", func() {
		m := baseModel(1)
		Expect(m.tickRate).To(Equal(50 * time.Millisecond))
	})

	It("defaults width to 80", func() {
		m := baseModel(1)
		Expect(m.width).To(Equal(80))
	})

	It("sets the root path", func() {
		m := baseModel(1)
		Expect(m.rootPath).To(Equal("/root"))
	})

	It("sets maxDepth", func() {
		m := baseModel(1)
		Expect(m.maxDepth).To(Equal(uint(5)))
	})

	It("stores the theme", func() {
		m := baseModel(1)
		Expect(m.theme).NotTo(Equal(contract.Theme{}))
	})

	It("initialises the counted map", func() {
		m := baseModel(1)
		Expect(m.counted).NotTo(BeNil())
		Expect(m.counted).To(HaveLen(0))
	})

	It("initialises the progress model", func() {
		m := baseModel(1)
		Expect(m.progress).NotTo(BeNil())
	})
})

// ---------------------------------------------------------------------------
// RegisterAll / Lookup Integration
// ---------------------------------------------------------------------------

var _ = Describe("RegisterAll smoke test", func() {
	It("Lookup returns all spinner types after RegisterAll", func() {
		traffic.RegisterAll()
		for name := range traffic.SpinnerNames {
			def, ok := traffic.Lookup(name)
			Expect(ok).To(BeTrue(), "Lookup(%q) should succeed", name)
			Expect(def.Frames).NotTo(BeNil())
			hasNonEmpty := def.Frames(0) != "" || def.Frames(1) != "" || def.Frames(2) != "" || def.Frames(3) != ""
			Expect(hasNonEmpty).To(BeTrue(), "spinner %q should have a non-empty frame within the first 4 ticks", name)
		}
	})
})

// ---------------------------------------------------------------------------
// Model.Init
// ---------------------------------------------------------------------------

var _ = Describe("Model.Init", func() {
	It("returns a tick command that produces a tickMsg", func() {
		m := baseModel(1)
		cmd := m.Init()
		Expect(cmd).NotTo(BeNil())

		msg := invokeCmd(cmd)
		_, ok := msg.(tickMsg)
		Expect(ok).To(BeTrue())
	})
})

// ---------------------------------------------------------------------------
// Model.Update — WindowSizeMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — WindowSizeMsg", func() {
	It("updates the width", func() {
		m := baseModel(1)
		updated, cmd := update(m, tea.WindowSizeMsg{Width: 120})
		Expect(cmd).To(BeNil())
		Expect(updated.width).To(Equal(120))
	})
})

// ---------------------------------------------------------------------------
// Model.Update — KeyMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — KeyMsg", func() {
	It("ignores any key before done", func() {
		m := baseModel(1)
		_, cmd := update(m, tea.KeyPressMsg{})
		Expect(cmd).To(BeNil())
	})

	It("ignores any key after done (non-space)", func() {
		m := baseModel(1)
		m.done = true
		_, cmd := update(m, tea.KeyPressMsg{})
		Expect(cmd).To(BeNil())
	})

	It("does not quit on space before done", func() {
		m := baseModel(1)
		_, cmd := update(m, tea.KeyPressMsg{})
		Expect(cmd).To(BeNil())
	})

	It("returns tea.Quit on ctrl+c before done", func() {
		m := baseModel(1)
		_, cmd := update(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		Expect(cmd).NotTo(BeNil())
	})

	It("returns tea.Quit on ctrl+c after done", func() {
		m := baseModel(1)
		m.done = true
		_, cmd := update(m, tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		Expect(cmd).NotTo(BeNil())
	})
})

// ---------------------------------------------------------------------------
// initLaneSkip
// ---------------------------------------------------------------------------

var _ = Describe("initLaneSkip", func() {
	It("returns zero factors when no lane has IntervalMs", func() {
		lanes := []Lane{
			{Emoji: "🔍", Label: "a", FrameFn: noopFrame},
			{Emoji: "🔍", Label: "b", FrameFn: noopFrame},
		}
		factors := initLaneSkip(lanes, 50*time.Millisecond)
		Expect(factors).To(HaveLen(2))
		Expect(factors[0]).To(Equal(0))
		Expect(factors[1]).To(Equal(0))
	})

	It("computes skip factor from IntervalMs / tickRate", func() {
		lanes := []Lane{
			{Emoji: "🔍", Label: "fast", FrameFn: noopFrame},
			{Emoji: "🔍", Label: "slow", FrameFn: noopFrame, IntervalMs: 5000},
			{Emoji: "🔍", Label: "medium", FrameFn: noopFrame, IntervalMs: 500},
		}
		// tickRate = 50ms → 5000/50 = 100, 500/50 = 10
		factors := initLaneSkip(lanes, 50*time.Millisecond)
		Expect(factors).To(HaveLen(3))
		Expect(factors[0]).To(Equal(0))
		Expect(factors[1]).To(Equal(100))
		Expect(factors[2]).To(Equal(10))
	})

	It("defaults tickMs to 50 when zero", func() {
		lanes := []Lane{
			{Emoji: "🔍", Label: "slow", FrameFn: noopFrame, IntervalMs: 500},
		}
		factors := initLaneSkip(lanes, 0)
		Expect(factors[0]).To(Equal(10))
	})

	It("clamps minimum skip factor to 1", func() {
		lanes := []Lane{
			{Emoji: "🔍", Label: "barely", FrameFn: noopFrame, IntervalMs: 25},
		}
		// 25/50 = 0 → clamped to 1
		factors := initLaneSkip(lanes, 50*time.Millisecond)
		Expect(factors[0]).To(Equal(1))
	})
})

// ---------------------------------------------------------------------------
// Model.Update — tickMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — tickMsg", func() {
	It("increments ticks on all lanes", func() {
		m := baseModel(2)
		Expect(m.lanes[0].tick).To(Equal(0))
		Expect(m.lanes[1].tick).To(Equal(0))

		updated, cmd := update(m, tickMsg(time.Now()))
		Expect(updated.lanes[0].tick).To(Equal(1))
		Expect(updated.lanes[1].tick).To(Equal(1))
		_ = cmd
	})

	It("reschedules the next tick when not done", func() {
		m := baseModel(1)
		_, cmd := update(m, tickMsg(time.Now()))
		Expect(cmd).NotTo(BeNil())

		msg := invokeCmd(cmd)
		_, ok := msg.(tickMsg)
		Expect(ok).To(BeTrue())
	})

	It("does not reschedule when done", func() {
		m := baseModel(1)
		m.done = true
		_, cmd := update(m, tickMsg(time.Now()))
		Expect(cmd).To(BeNil())
	})

	It("does not start the timer in realMode when start is zero", func() {
		m := baseModel(1)
		m.realMode = true
		updated, cmd := update(m, tickMsg(time.Now()))
		Expect(updated.start.IsZero()).To(BeTrue())
		_ = cmd
	})

	It("starts the timer in demo mode when start is zero", func() {
		m := baseModel(1)
		m.realMode = false
		Expect(m.start.IsZero()).To(BeTrue())

		updated, cmd := update(m, tickMsg(time.Now()))
		Expect(updated.start.IsZero()).To(BeFalse())
		_ = cmd
	})

	It("advances slow lane fewer ticks than fast lane based on IntervalMs", func() {
		fast := Lane{Emoji: "🔍", Label: "fast", FrameFn: noopFrame}
		slow := Lane{Emoji: "🐢", Label: "slow", FrameFn: noopFrame, IntervalMs: 500}
		lanes := []Lane{fast, slow}
		m := NewModel(lanes, 50*time.Millisecond, "/root", 5, testTheme(), false)

		// After 10 ticks at 50ms:
		//   fast (skip=0) → tick = 10
		//   slow (IntervalMs=500, skip=10) → tick = 1 (every 10th tick)
		for range 10 {
			m, _ = update(m, tickMsg(time.Now()))
		}
		Expect(m.lanes[0].tick).To(Equal(10), "fast lane should advance every tick")
		Expect(m.lanes[1].tick).To(Equal(1), "slow lane should advance every 10th tick")
	})

	It("respects different intervals per lane", func() {
		fast := Lane{Emoji: "🔍", Label: "fast", FrameFn: noopFrame}
		medium := Lane{Emoji: "⚡", Label: "medium", FrameFn: noopFrame, IntervalMs: 200}
		verySlow := Lane{Emoji: "🐌", Label: "very-slow", FrameFn: noopFrame, IntervalMs: 5000}
		lanes := []Lane{fast, medium, verySlow}
		m := NewModel(lanes, 50*time.Millisecond, "/root", 5, testTheme(), false)

		for range 100 {
			m, _ = update(m, tickMsg(time.Now()))
		}
		Expect(m.lanes[0].tick).To(Equal(100), "fast: every tick")
		Expect(m.lanes[1].tick).To(Equal(25), "medium: 200/50=4, 100/4=25")
		Expect(m.lanes[2].tick).To(Equal(1), "very-slow: 5000/50=100, 100/100=1")
	})
})

// ---------------------------------------------------------------------------
// Model.Update — OvertureMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — OvertureMsg", func() {
	It("sets rootPath, realMode, start time and pipelineName", func() {
		m := baseModel(1)
		Expect(m.realMode).To(BeFalse())
		Expect(m.start.IsZero()).To(BeTrue())

		updated, cmd := update(m, OvertureMsg{
			Root:         "/project/src",
			Caption:      "files",
			ActionName:   "build",
			PipelineName: "ci",
			DateFormat:   "Mon, 02 Jan 2006 15:04:05 MST",
		})
		Expect(cmd).To(BeNil())

		Expect(updated.rootPath).To(Equal("/project/src"))
		Expect(updated.realMode).To(BeTrue())
		Expect(updated.start.IsZero()).To(BeFalse())
		Expect(updated.pipelineName).To(Equal("ci"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update — CensusMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — CensusMsg", func() {
	It("sets totalFiles and totalDirs", func() {
		m := baseModel(1)
		updated, cmd := update(m, CensusMsg{TotalFiles: 100, TotalDirs: 20})
		Expect(cmd).To(BeNil())
		Expect(updated.totalFiles).To(Equal(uint(100)))
		Expect(updated.totalDirs).To(Equal(uint(20)))
	})

	It("increases maxDepth when a larger value is seen", func() {
		m := baseModel(1)
		Expect(m.maxDepth).To(Equal(uint(5)))

		updated, cmd := update(m, CensusMsg{MaxDepth: 10})
		Expect(cmd).To(BeNil())
		Expect(updated.maxDepth).To(Equal(uint(10)))
	})

	It("does not decrease maxDepth", func() {
		m := baseModel(1)
		Expect(m.maxDepth).To(Equal(uint(5)))

		updated, cmd := update(m, CensusMsg{MaxDepth: 3})
		Expect(cmd).To(BeNil())
		Expect(updated.maxDepth).To(Equal(uint(5)))
	})
})

// ---------------------------------------------------------------------------
// Model.Update — MotifMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — MotifMsg", func() {
	It("updates the current lane and advances the index round-robin", func() {
		m := baseModel(2)
		Expect(m.currentLaneIdx).To(Equal(0))

		first := MotifMsg{Data: MotifData{
			Path: "/root/a.txt", Name: "a.txt", IsDir: false, Depth: 1,
		}}
		updated1, cmd := update(m, first)
		Expect(cmd).To(BeNil())
		Expect(updated1.lanes[0].Path).To(Equal("/root/a.txt"))
		Expect(updated1.lanes[0].Name).To(Equal("a.txt"))
		Expect(updated1.lanes[0].IsDir).To(BeFalse())
		Expect(updated1.lanes[0].Depth).To(Equal(uint(1)))
		Expect(updated1.currentLaneIdx).To(Equal(1))

		second := MotifMsg{Data: MotifData{
			Path: "/root/b.txt", Name: "b.txt", IsDir: false, Depth: 2,
		}}
		updated2, cmd := update(updated1, second)
		Expect(cmd).To(BeNil())
		Expect(updated2.lanes[1].Path).To(Equal("/root/b.txt"))
		Expect(updated2.currentLaneIdx).To(Equal(0))

		// Third wraps around to lane 0
		third := MotifMsg{Data: MotifData{
			Path: "/root/c.txt", Name: "c.txt", IsDir: false, Depth: 3,
		}}
		updated3, cmd := update(updated2, third)
		Expect(cmd).To(BeNil())
		Expect(updated3.lanes[0].Path).To(Equal("/root/c.txt"))
		Expect(updated3.currentLaneIdx).To(Equal(1))
	})

	It("counts each unique path once for progress", func() {
		m := baseModel(1)
		m.totalFiles = 10
		Expect(m.files).To(Equal(0))

		// First unique path
		updated, cmd := update(m, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(cmd).To(BeNil())
		Expect(updated.files).To(Equal(1))

		// Same path again — not counted
		updated, cmd = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(cmd).To(BeNil())
		Expect(updated.files).To(Equal(1))

		// Different path — counted
		updated, cmd = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/b.txt", IsDir: false,
		}})
		Expect(cmd).To(BeNil())
		Expect(updated.files).To(Equal(2))
	})

	It("increments dirs for directories", func() {
		m := baseModel(1)
		Expect(m.dirs).To(Equal(0))

		updated, cmd := update(m, MotifMsg{Data: MotifData{
			Path: "/root/src", IsDir: true,
		}})
		Expect(cmd).To(BeNil())
		Expect(updated.dirs).To(Equal(1))
		Expect(updated.files).To(Equal(0))
	})

	It("sets action and pipeline info on the lane", func() {
		m := baseModel(1)

		updated, cmd := update(m, MotifMsg{Data: MotifData{
			Path: "/root/f.txt", ActionName: "encode", PipelineName: "pipe",
			CommandOutput: "ok", ExecutionString: "ffmpeg", DryRun: true,
			Err: errors.New("boom"),
		}})
		Expect(cmd).To(BeNil())

		Expect(updated.lanes[0].ActionName).To(Equal("encode"))
		Expect(updated.lanes[0].PipelineName).To(Equal("pipe"))
		Expect(updated.lanes[0].CommandOutput).To(Equal("ok"))
		Expect(updated.lanes[0].ExecutionString).To(Equal("ffmpeg"))
		Expect(updated.lanes[0].DryRun).To(BeTrue())
		Expect(updated.lanes[0].Err).To(MatchError("boom"))
	})

	It("computes percent when totalFiles > 0", func() {
		m := baseModel(1)
		m.totalFiles = 10

		updated, cmd := update(m, MotifMsg{Data: MotifData{
			Path: "/root/1.txt", IsDir: false,
		}})
		Expect(cmd).To(BeNil())
		Expect(updated.percent).To(Equal(10)) // 1/10 = 10%

		updated, cmd = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/2.txt", IsDir: false,
		}})
		Expect(cmd).To(BeNil())
		Expect(updated.percent).To(Equal(20)) // 2/10 = 20%
	})

	It("handles empty lanes gracefully", func() {
		m := baseModel(0)

		_, cmd := update(m, MotifMsg{Data: MotifData{
			Path: "/root/f.txt",
		}})
		Expect(cmd).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// Model.Update — CompleteMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — CompleteMsg", func() {
	It("marks the model as done", func() {
		m := baseModel(1)
		Expect(m.done).To(BeFalse())

		updated, cmd := update(m, CompleteMsg{})
		Expect(cmd).To(BeNil())
		Expect(updated.done).To(BeTrue())
	})

	It("sets files, dirs, errors and elapsed", func() {
		m := baseModel(1)

		updated, cmd := update(m, CompleteMsg{
			Files: 42, Dirs: 7, Elapsed: 5 * time.Second,
		})
		Expect(cmd).To(BeNil())

		Expect(updated.files).To(Equal(42))
		Expect(updated.dirs).To(Equal(7))
		Expect(updated.errors).To(Equal(0))
		Expect(updated.elapsed).To(Equal(5 * time.Second))
	})

	It("captures the first error message", func() {
		m := baseModel(1)

		updated, cmd := update(m, CompleteMsg{
			Errs: []error{
				errors.New("first error"),
				errors.New("second error"),
			},
		})
		Expect(cmd).To(BeNil())

		Expect(updated.errMsg).To(Equal("first error"))
		Expect(updated.errors).To(Equal(2))
	})

	It("recomputes percent when totalFiles > 0", func() {
		m := baseModel(1)
		m.totalFiles = 100

		updated, cmd := update(m, CompleteMsg{Files: 80})
		Expect(cmd).To(BeNil())
		Expect(updated.percent).To(Equal(80))
	})
})

// ---------------------------------------------------------------------------
// Model.Update — default case (unknown message)
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update — unknown message", func() {
	It("returns the model unchanged with nil cmd", func() {
		m := baseModel(1)
		_, cmd := update(m, "some-unknown-msg")
		Expect(cmd).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// RenderLandingStrip (via widget)
// ---------------------------------------------------------------------------

var _ = Describe("RenderLandingStrip", func() {
	It("returns the command output wrapped in branch/landing-strip styles", func() {
		m := baseModel(1)
		styles := widget.LandingStripStyles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := widget.RenderLandingStrip(widget.LandingStripConfig{
			CommandOutput: "ffmpeg -i input.mp4",
		}, styles)
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(ContainSubstring("ffmpeg -i input.mp4"))
	})

	It("returns the execution string when DryRun is true", func() {
		m := baseModel(1)
		styles := widget.LandingStripStyles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := widget.RenderLandingStrip(widget.LandingStripConfig{
			CommandOutput:   "real command",
			ExecutionString: "dry-run command",
			DryRun:          true,
		}, styles)
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(ContainSubstring("dry-run command"))
		Expect(result).NotTo(ContainSubstring("real command"))
	})

	It("returns empty string when CommandOutput is empty and not DryRun", func() {
		m := baseModel(1)
		styles := widget.LandingStripStyles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := widget.RenderLandingStrip(widget.LandingStripConfig{}, styles)
		Expect(result).To(BeEmpty())
	})

	It("returns empty string when ExecutionString is empty and DryRun is true", func() {
		m := baseModel(1)
		styles := widget.LandingStripStyles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := widget.RenderLandingStrip(widget.LandingStripConfig{
			DryRun: true,
		}, styles)
		Expect(result).To(BeEmpty())
	})
})

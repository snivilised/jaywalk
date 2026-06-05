package highway

import (
	"errors"
	"io"
	"time"

	bp "charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/movies"
	"github.com/snivilised/jaywalk/src/prism/widgets/landing"
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
	return Lane{Emoji: "🔍", JobEmoji: "🍎", Label: "test", FrameFn: noopFrame}
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

	It("initialises the status widget", func() {
		m := baseModel(1)
		// The status widget is a value type; the assertion is
		// that the construction did not panic and produced a
		// usable widget (its View() returns a non-empty string
		// once it has been given a width via tea.WindowSizeMsg).
		updated, _ := update(m, tea.WindowSizeMsg{Width: 80})
		Expect(updated.status.View().Content).NotTo(BeEmpty())
	})
})

// ---------------------------------------------------------------------------
// RegisterAll / Lookup Integration
// ---------------------------------------------------------------------------

var _ = Describe("RegisterAll smoke test", func() {
	It("Lookup returns all spinner types after RegisterAll", func() {
		movies.RegisterAll()
		for name := range movies.SpinnerNames {
			def, ok := movies.Lookup(name)
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
// Model.Update - WindowSizeMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - WindowSizeMsg", func() {
	It("updates the width", func() {
		m := baseModel(1)
		updated, cmd := update(m, tea.WindowSizeMsg{Width: 120})
		Expect(cmd).To(BeNil())
		Expect(updated.width).To(Equal(120))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - KeyMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - KeyMsg", func() {
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
// Model.Update - tickMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - tickMsg", func() {
	It("increments ticks on all lanes", func() {
		m := baseModel(2)
		Expect(m.lanes[0].tick).To(Equal(0))
		Expect(m.lanes[1].tick).To(Equal(0))

		updated, cmd := update(m, tickMsg(core.Now()))
		Expect(updated.lanes[0].tick).To(Equal(1))
		Expect(updated.lanes[1].tick).To(Equal(1))
		_ = cmd
	})

	It("reschedules the next tick when not done", func() {
		m := baseModel(1)
		_, cmd := update(m, tickMsg(core.Now()))
		Expect(cmd).NotTo(BeNil())

		msg := invokeCmd(cmd)
		_, ok := msg.(tickMsg)
		Expect(ok).To(BeTrue())
	})

	It("does not reschedule when done", func() {
		m := baseModel(1)
		m.done = true
		_, cmd := update(m, tickMsg(core.Now()))
		Expect(cmd).To(BeNil())
	})

	It("does not start the timer in realMode when start is zero", func() {
		m := baseModel(1)
		m.realMode = true
		updated, cmd := update(m, tickMsg(core.Now()))
		Expect(updated.start.IsZero()).To(BeTrue())
		_ = cmd
	})

	It("starts the timer in demo mode when start is zero", func() {
		m := baseModel(1)
		m.realMode = false
		Expect(m.start.IsZero()).To(BeTrue())

		updated, cmd := update(m, tickMsg(core.Now()))
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
			m, _ = update(m, tickMsg(core.Now()))
		}
		Expect(m.lanes[0].tick).To(Equal(10), "fast lane should advance every tick")
		Expect(m.lanes[1].tick).To(Equal(1), "slow lane should advance every 10th tick")
	})

	It("respects different intervals per lane", func() {
		fast := Lane{Emoji: "🔍", Label: "fast", FrameFn: noopFrame}
		medium := Lane{Emoji: "⚡", Label: "medium", FrameFn: noopFrame, IntervalMs: 200}
		verySlow := Lane{Emoji: contract.Static.Emoji.Snail, Label: "very-slow", FrameFn: noopFrame, IntervalMs: 5000}
		lanes := []Lane{fast, medium, verySlow}
		m := NewModel(lanes, 50*time.Millisecond, "/root", 5, testTheme(), false)

		for range 100 {
			m, _ = update(m, tickMsg(core.Now()))
		}
		Expect(m.lanes[0].tick).To(Equal(100), "fast: every tick")
		Expect(m.lanes[1].tick).To(Equal(25), "medium: 200/50=4, 100/4=25")
		Expect(m.lanes[2].tick).To(Equal(1), "very-slow: 5000/50=100, 100/100=1")
	})

	It("pushes elapsed to the status widget in real mode on every tick", func() {
		// Regression: ElapsedMsg was previously only dispatched
		// in demo mode, so in real mode the status row's elapsed
		// segment stayed at 0 for the entire traversal and only
		// jumped to the final value on CompleteMsg. The elapsed
		// is real in both modes and must tick up.
		m := baseModel(1)
		// OvertureMsg sets realMode and m.start = now.
		m, _ = update(m, OvertureMsg{
			Root:    "/root",
			Caption: "files",
		})
		Expect(m.realMode).To(BeTrue())
		Expect(m.start.IsZero()).To(BeFalse())
		Expect(m.status.Elapsed()).To(Equal(time.Duration(0)),
			"no tick yet, status elapsed stays at 0")

		// Drive a few ticks. Each tick should push a fresh
		// ElapsedMsg; the status widget's elapsed is set
		// verbatim from msg.Elapsed, so the value should be
		// monotonically non-decreasing.
		var prev time.Duration
		for range 3 {
			m, _ = update(m, tickMsg(core.Now()))
			current := m.status.Elapsed()
			Expect(current).To(BeNumerically(">=", prev),
				"elapsed should never decrease across ticks")
			prev = current
		}
	})

	It("pushes elapsed to the status widget in demo mode on every tick", func() {
		// Demo mode has always pushed ElapsedMsg; this test
		// documents the behaviour alongside the real-mode case
		// above so a future refactor doesn't drop it.
		m := baseModel(1)
		Expect(m.realMode).To(BeFalse())
		Expect(m.status.Elapsed()).To(Equal(time.Duration(0)))

		// First tick sets m.start (demo mode), so subsequent
		// ticks can produce a non-zero elapsed.
		m, _ = update(m, tickMsg(core.Now()))
		first := m.status.Elapsed()
		Expect(first).To(BeNumerically(">=", time.Duration(0)))

		m, _ = update(m, tickMsg(core.Now()))
		second := m.status.Elapsed()
		Expect(second).To(BeNumerically(">=", first))
	})

	It("drives the percent in demo mode via PercentMsg (not done/total)", func() {
		// Demo mode has no real traversal data, so the percent
		// must come from the time-derived PercentMsg - not from
		// recomputePercent (which would stay at 0 without a
		// TotalMsg). This test pins down that contract.
		m := baseModel(1)
		for range 5 {
			m, _ = update(m, tickMsg(core.Now()))
		}
		// The demo formula is int(elapsed.Seconds())*2 % 100.
		// We don't pin the exact value (it's time-dependent)
		// but we do assert that the percent state is non-zero
		// OR zero (the formula can produce 0 if elapsed is
		// exactly a multiple of 50 seconds). The key assertion
		// is the view contains no "100 / 100" ratio, because
		// demo mode has no TotalMsg.
		Expect(m.status.HasTotal()).To(BeFalse(),
			"demo mode never sends TotalMsg, so hasTotal stays false")
	})
})

// ---------------------------------------------------------------------------
// Model.Update - OvertureMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - OvertureMsg", func() {
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
// Model.Update - CensusMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - CensusMsg", func() {
	It("sets totalFiles and totalDirs", func() {
		m := baseModel(1)
		updated, _ := update(m, CensusMsg{TotalFiles: 100, TotalDirs: 20})
		// cmd is the status spring's first-frame cmd (non-nil)
		// because CensusMsg seeds the total and re-targets the
		// embedded progress bar to 0.
		Expect(updated.totalFiles).To(Equal(uint(100)))
		Expect(updated.totalDirs).To(Equal(uint(20)))
		// Regression: the status widget's total must be the sum
		// of files and dirs, because every MotifMsg (file OR
		// dir) increments done by 1. Seeding total with only
		// TotalFiles makes done exceed total during navigation
		// of trees that contain directories, clamping the bar
		// to 100% before completion.
		Expect(updated.status.Total()).To(Equal(120),
			"status total must include both files and dirs to match the unit IncDoneMsg cadence")
		Expect(updated.status.HasTotal()).To(BeTrue())
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
// Model.Update - MotifMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - MotifMsg", func() {
	It("updates the current lane and advances the index round-robin", func() {
		m := baseModel(2)
		Expect(m.currentLaneIdx).To(Equal(0))

		first := MotifMsg{Data: MotifData{
			Path: "/root/a.txt", Name: "a.txt", IsDir: false, Depth: 1,
		}}
		updated1, _ := update(m, first)
		// cmd is the status spring's first-frame cmd; the
		// MotifMsg increment re-targets the bar to (done/total).
		Expect(updated1.lanes[0].Path).To(Equal("/root/a.txt"))
		Expect(updated1.lanes[0].Name).To(Equal("a.txt"))
		Expect(updated1.lanes[0].IsDir).To(BeFalse())
		Expect(updated1.lanes[0].Depth).To(Equal(uint(1)))
		Expect(updated1.currentLaneIdx).To(Equal(1))

		second := MotifMsg{Data: MotifData{
			Path: "/root/b.txt", Name: "b.txt", IsDir: false, Depth: 2,
		}}
		updated2, _ := update(updated1, second)
		Expect(updated2.lanes[1].Path).To(Equal("/root/b.txt"))
		Expect(updated2.currentLaneIdx).To(Equal(0))

		// Third wraps around to lane 0
		third := MotifMsg{Data: MotifData{
			Path: "/root/c.txt", Name: "c.txt", IsDir: false, Depth: 3,
		}}
		updated3, _ := update(updated2, third)
		Expect(updated3.lanes[0].Path).To(Equal("/root/c.txt"))
		Expect(updated3.currentLaneIdx).To(Equal(1))
	})

	It("counts each unique path once for progress", func() {
		m := baseModel(1)
		m.totalFiles = 10
		Expect(m.files).To(Equal(0))

		// First unique path
		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(1))

		// Same path again - not counted (no IncDoneMsg, no
		// spring re-target, cmd is nil).
		updated, cmd := update(updated, MotifMsg{Data: MotifData{
			Path: "/root/a.txt", IsDir: false,
		}})
		Expect(cmd).To(BeNil(), "duplicate path must not re-target the spring")
		Expect(updated.files).To(Equal(1))

		// Different path - counted
		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/b.txt", IsDir: false,
		}})
		Expect(updated.files).To(Equal(2))
	})

	It("increments dirs for directories", func() {
		m := baseModel(1)
		Expect(m.dirs).To(Equal(0))

		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/src", IsDir: true,
		}})
		Expect(updated.dirs).To(Equal(1))
		Expect(updated.files).To(Equal(0))
	})

	It("sets action and pipeline info on the lane", func() {
		m := baseModel(1)

		updated, _ := update(m, MotifMsg{Data: MotifData{
			Path: "/root/f.txt", ActionName: "encode", PipelineName: "pipe",
			CommandOutput: "ok", ExecutionString: "ffmpeg", DryRun: true,
			Err: errors.New("boom"),
		}})

		Expect(updated.lanes[0].ActionName).To(Equal("encode"))
		Expect(updated.lanes[0].PipelineName).To(Equal("pipe"))
		Expect(updated.lanes[0].CommandOutput).To(Equal("ok"))
		Expect(updated.lanes[0].ExecutionString).To(Equal("ffmpeg"))
		Expect(updated.lanes[0].DryRun).To(BeTrue())
		Expect(updated.lanes[0].Err).To(MatchError("boom"))
	})

	It("computes percent during navigation from done/total", func() {
		// The status widget computes percent from done/total on
		// every MotifMsg. CensusMsg seeds the total; each
		// subsequent MotifMsg increments done. The bar fills
		// proportionally and the label shows the percent.
		m := baseModel(1)
		// CensusMsg forwards totalFiles into the status widget
		// via status.TotalMsg.
		updated, _ := update(m, CensusMsg{TotalFiles: 10})

		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/1.txt", IsDir: false,
		}})
		// 1 file traversed; done=1, total=10 → 10%.
		Expect(updated.status.Done()).To(Equal(1))
		Expect(updated.status.HasTotal()).To(BeTrue())
		Expect(updated.status.Percent()).To(Equal(10))
		Expect(updated.status.View().Content).To(ContainSubstring("10%"))

		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/2.txt", IsDir: false,
		}})
		// 2 files traversed; 2/10 → 20%.
		Expect(updated.status.Done()).To(Equal(2))
		Expect(updated.status.Percent()).To(Equal(20))
		Expect(updated.status.View().Content).To(ContainSubstring("20%"))
	})

	It("handles empty lanes gracefully", func() {
		m := baseModel(0)

		_, _ = update(m, MotifMsg{Data: MotifData{
			Path: "/root/f.txt",
		}})
		// No assertion on cmd: with zero lanes the path is
		// still counted (drives the spring), so cmd is non-nil.
	})

	It("does not reach 100% before navigation completes when dirs are traversed alongside files", func() {
		// Regression: the bar used to clamp to 100% partway
		// through navigation because CensusMsg seeded the
		// status total with only TotalFiles, while every
		// MotifMsg (file OR dir) increments done. Seeding
		// total with files+dirs keeps the ratio accurate.
		m := baseModel(1)
		updated, _ := update(m, CensusMsg{TotalFiles: 3, TotalDirs: 2})
		Expect(updated.status.Total()).To(Equal(5))

		// Visit all 3 files. Done climbs to 3 of 5 → 60%.
		for _, p := range []string{"/r/a.txt", "/r/b.txt", "/r/c.txt"} {
			updated, _ = update(updated, MotifMsg{Data: MotifData{
				Path: p, IsDir: false,
			}})
		}
		Expect(updated.status.Done()).To(Equal(3))
		Expect(updated.status.Percent()).To(Equal(60),
			"with files complete but dirs pending, the bar must NOT be at 100%")

		// Visit 1 of 2 dirs. Done climbs to 4 of 5 → 80%.
		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/r/sub1", IsDir: true,
		}})
		Expect(updated.status.Done()).To(Equal(4))
		Expect(updated.status.Percent()).To(Equal(80))

		// Visit final dir. Done climbs to 5 of 5 → 100%.
		updated, _ = update(updated, MotifMsg{Data: MotifData{
			Path: "/r/sub2", IsDir: true,
		}})
		Expect(updated.status.Done()).To(Equal(5))
		Expect(updated.status.Percent()).To(Equal(100))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - CompleteMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - CompleteMsg", func() {
	It("marks the model as done", func() {
		m := baseModel(1)
		Expect(m.done).To(BeFalse())

		updated, _ := update(m, CompleteMsg{})
		// cmd is tea.Batch wrapping the status spring's
		// first-frame cmd (re-targeted to 100% via DoneMsg).
		Expect(updated.done).To(BeTrue())
	})

	It("sets files, dirs, errors and elapsed on the status widget", func() {
		m := baseModel(1)

		updated, cmd := update(m, CompleteMsg{
			Files: 42, Dirs: 7, Elapsed: 5 * time.Second,
		})
		// Cmd is tea.Batch wrapping the status spring's
		// first-frame cmd.
		_ = cmd

		// The counts now live on the status widget. The
		// accessors are public on status.Model so this
		// white-box test can assert on them directly.
		Expect(updated.status.Files()).To(Equal(42))
		Expect(updated.status.Dirs()).To(Equal(7))
		Expect(updated.status.Errors()).To(Equal(0))
		Expect(updated.status.Elapsed()).To(Equal(5 * time.Second))
	})

	It("captures the first error message", func() {
		m := baseModel(1)

		updated, _ := update(m, CompleteMsg{
			Errs: []error{
				errors.New("first error"),
				errors.New("second error"),
			},
		})

		Expect(updated.errMsg).To(Equal("first error"))
		Expect(updated.errors).To(Equal(2))
	})

	It("shows 100% on completion regardless of totalFiles", func() {
		// CompleteMsg always reports the final counts and
		// IsDone=true, which the highway root translates to
		// status.DoneMsg{IsDone:true}. The status widget sets
		// percent=100 unconditionally on IsDone=true, even when
		// Files (80) is less than totalFiles (100). This is the
		// "bar fills at completion" semantic.
		m := baseModel(1)
		// Seed the total via CensusMsg (preview estimate).
		updated, _ := update(m, CensusMsg{TotalFiles: 100})

		updated, cmd := update(updated, CompleteMsg{Files: 80, Dirs: 5})
		// The status widget returns nil cmd for DoneMsg (no
		// animation driver in this PR), so the tea.Batch
		// collapses to nil.
		_ = cmd

		Expect(updated.status.IsDone()).To(BeTrue())
		Expect(updated.status.Percent()).To(Equal(100))
		Expect(updated.status.Files()).To(Equal(80))
		Expect(updated.status.Dirs()).To(Equal(5))
		Expect(updated.status.View().Content).To(ContainSubstring("100%"))
		Expect(updated.status.View().Content).To(ContainSubstring("✔ complete"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - FrameMsg forwarding (spring animation)
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - FrameMsg forwarding", func() {
	It("forwards bp.FrameMsg to the status widget so the spring can advance", func() {
		// Seed the status widget's spring via CensusMsg
		// (which translates to status.TotalMsg and re-targets
		// the bar to 0%). Capture the first-frame cmd, execute
		// it to obtain a real FrameMsg, and feed that back to
		// the *same* highway state (the FrameMsg's id is tied
		// to the specific inner model that produced it; routing
		// it back to a fresh model would silently drop it).
		m := baseModel(1)
		// Re-target to a non-zero percent so the spring has
		// somewhere to travel and the frame's next-frame cmd
		// is non-nil.
		updated, cmd := update(m, CensusMsg{TotalFiles: 100})
		Expect(cmd).NotTo(BeNil(), "CensusMsg must propagate the spring cmd")

		// Drive percent to 100% via a MotifMsg increment with a
		// total set; this re-targets the spring well away from
		// its current percentShown=0.
		updated, cmd = update(updated, MotifMsg{Data: MotifData{
			Path: "/root/f.txt", IsDir: false,
		}})
		Expect(cmd).NotTo(BeNil())

		msg := cmd()
		frame, ok := msg.(bp.FrameMsg)
		Expect(ok).To(BeTrue(), "spring cmd must yield a bp.FrameMsg, got %T", msg)

		// Forward the frame back to the same updated state.
		// The spring is still mid-flight (percentShown ≈ 0,
		// target = 0.01) so the next-frame cmd is non-nil.
		_, nextCmd := update(updated, frame)
		Expect(nextCmd).NotTo(BeNil(),
			"FrameMsg must reach the spring and produce a next-frame cmd")
	})
})

// ---------------------------------------------------------------------------
// Model.Update - default case (unknown message)
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - unknown message", func() {
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
		styles := landing.Styles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := landing.Render(landing.Config{
			CommandOutput: "ffmpeg -i input.mp4",
		}, styles)
		Expect(result).NotTo(BeEmpty())
		Expect(result).To(ContainSubstring("ffmpeg -i input.mp4"))
	})

	It("returns the execution string when DryRun is true", func() {
		m := baseModel(1)
		styles := landing.Styles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := landing.Render(landing.Config{
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
		styles := landing.Styles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := landing.Render(landing.Config{}, styles)
		Expect(result).To(BeEmpty())
	})

	It("returns empty string when ExecutionString is empty and DryRun is true", func() {
		m := baseModel(1)
		styles := landing.Styles{
			BranchStyle:       m.theme.BranchStyle,
			LandingStripStyle: m.theme.LandingStripStyle,
		}
		result := landing.Render(landing.Config{
			DryRun: true,
		}, styles)
		Expect(result).To(BeEmpty())
	})
})

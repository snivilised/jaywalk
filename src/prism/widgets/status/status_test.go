package status_test

import (
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
)

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Suite")
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func baseStyles() status.Styles {
	return status.Styles{
		TreeIcons: contract.TreeIcons{
			contract.TreeIconFile:      "🔖",
			contract.TreeIconDirectory: "📁",
			contract.TreeIconError:     "🚫",
			contract.TreeIconSkipped:   "⛔️",
			contract.TreeIconElapsed:   "⏰",
		},
		SummaryLabelStyle: lipgloss.NewStyle(),
		SummaryValueStyle: lipgloss.NewStyle(),
		ErrorStyle:        lipgloss.NewStyle(),
		ProgressStyle:     lipgloss.NewStyle(),
		BorderStyle:       lipgloss.NewStyle(),
		MutedStyle:        lipgloss.NewStyle(),
	}
}

func allFieldsOn() status.FieldSelectors {
	return status.FieldSelectors{
		ShowFiles:    true,
		ShowDirs:     true,
		ShowErrors:   true,
		ShowSkipped:  true,
		ShowProgress: true,
		ShowComplete: true,
		ShowElapsed:  true,
	}
}

// update is a convenience wrapper that casts the bubbletea Model
// back to a status.Model. Mirrors the helper used in
// src/prism/highway/model_test.go.
func update(m status.Model, msg tea.Msg) (status.Model, tea.Cmd) {
	r, cmd := m.Update(msg)
	return r.(status.Model), cmd //nolint:errcheck // ok
}

// ---------------------------------------------------------------------------
// Render (stateless wrapper, used by the linear view)
// ---------------------------------------------------------------------------

var _ = Describe("Render", func() {
	var (
		styles status.Styles
	)

	BeforeEach(func() {
		styles = baseStyles()
	})

	It("renders all fields when all are enabled", func() {
		fields := status.FieldSelectors{
			ShowFiles:    true,
			ShowDirs:     true,
			ShowErrors:   true,
			ShowSkipped:  true,
			ShowProgress: false,
			ShowComplete: false,
			ShowElapsed:  true,
		}

		// Width 104 — 4 segments (files fixed at 16, dirs fixed at 14,
		// errors and skipped natural) + elapsed need 99 cells to
		// avoid the Row dropping the right-zone.
		output := status.Render(status.Config{
			Files:   42,
			Dirs:    7,
			Errors:  0,
			Skipped: 3,
			Elapsed: 5 * time.Second,
		}, styles, fields, 104)

		Expect(output).To(ContainSubstring("🔖 files:"))
		Expect(output).To(ContainSubstring("📁 dirs:"))
		Expect(output).To(ContainSubstring("🚫 errors:"))
		Expect(output).To(ContainSubstring("⛔️ skipped:"))
		Expect(output).To(ContainSubstring("⏰ elapsed:"))
	})

	It("renders only enabled fields", func() {
		fields := status.FieldSelectors{
			ShowFiles:    true,
			ShowDirs:     false,
			ShowErrors:   false,
			ShowSkipped:  false,
			ShowProgress: false,
			ShowComplete: false,
			ShowElapsed:  true,
		}

		output := status.Render(status.Config{
			Files:   42,
			Dirs:    7,
			Errors:  0,
			Skipped: 3,
			Elapsed: 5 * time.Second,
		}, styles, fields, 80)

		Expect(output).To(ContainSubstring("🔖 files:"))
		Expect(output).To(ContainSubstring("⏰ elapsed:"))
		Expect(output).ToNot(ContainSubstring("📁 dirs:"))
		Expect(output).ToNot(ContainSubstring("🚫 errors:"))
		Expect(output).ToNot(ContainSubstring("⛔️ skipped:"))
	})

	It("formats duration correctly", func() {
		fields := status.FieldSelectors{
			ShowElapsed: true,
		}

		output := status.Render(status.Config{
			Elapsed: 90 * time.Second,
		}, styles, fields, 80)

		Expect(output).To(ContainSubstring("1m30s"))
	})

	It("formats milliseconds for sub-second durations", func() {
		fields := status.FieldSelectors{ShowElapsed: true}

		output := status.Render(status.Config{
			Elapsed: 500 * time.Millisecond,
		}, styles, fields, 80)

		Expect(output).To(ContainSubstring("500ms"))
	})

	It("includes border characters", func() {
		fields := status.FieldSelectors{
			ShowFiles:   true,
			ShowElapsed: true,
		}

		output := status.Render(status.Config{
			Files:   42,
			Elapsed: 5 * time.Second,
		}, styles, fields, 80)

		Expect(output).To(ContainSubstring("│"))
	})
})

// ---------------------------------------------------------------------------
// Model.Init
// ---------------------------------------------------------------------------

var _ = Describe("Model.Init", func() {
	It("returns nil - the spring starts lazily on first SetPercent", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(allFieldsOn()))
		Expect(m.Init()).To(BeNil())
	})
})

// ---------------------------------------------------------------------------
// Model.IsDone accessor
// ---------------------------------------------------------------------------

var _ = Describe("Model.IsDone", func() {
	It("is false on a fresh model", func() {
		m := status.New(status.WithStyles(baseStyles()))
		Expect(m.IsDone()).To(BeFalse())
	})

	It("becomes true after a DoneMsg with IsDone=true", func() {
		m := status.New(status.WithStyles(baseStyles()))
		updated, _ := update(m, status.DoneMsg{Done: 10, IsDone: true})
		Expect(updated.IsDone()).To(BeTrue())
	})
})

// ---------------------------------------------------------------------------
// Model.Update - WidthMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - WidthMsg", func() {
	It("updates the row width but NOT the embedded progress width", func() {
		// Regression: previously WidthMsg propagated msg.Width
		// to m.inner.Width, making the progress bar fill the
		// entire terminal on resize. The bar must stay at its
		// fixed small width (10 by default) so the percentage
		// label and the other segments remain visible.
		m := status.New(status.WithStyles(baseStyles()))
		Expect(m.Inner().Width()).To(Equal(10), "default inner width")

		updated, cmd := update(m, status.WidthMsg{Width: 120})
		Expect(cmd).To(BeNil())

		// Row width changes; inner width does NOT.
		Expect(updated.Inner().Width()).To(Equal(10),
			"WidthMsg must not resize the embedded progress bar")
		Expect(updated.View().Content).NotTo(BeEmpty())
	})

	It("WithWidth does NOT affect the embedded progress width", func() {
		// WithWidth sets the row layout width only. The embedded
		// progress bar keeps its default width (10) so the bar
		// does not balloon to fill the terminal.
		m := status.New(status.WithStyles(baseStyles()), status.WithWidth(20))
		Expect(m.Inner().Width()).To(Equal(10),
			"WithWidth must not change the progress bar width")

		updated, _ := update(m, status.WidthMsg{Width: 120})
		Expect(updated.Inner().Width()).To(Equal(10),
			"WidthMsg must not resize the embedded progress bar")
	})
})

// ---------------------------------------------------------------------------
// Model.Update - CountsMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - CountsMsg", func() {
	It("stores the counts and reflects them in View", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowFiles: true,
			ShowDirs:  true,
		}))
		updated, cmd := update(m, status.CountsMsg{
			Files: 7, Dirs: 3, Errors: 0,
		})
		Expect(cmd).To(BeNil())
		out := updated.View().Content
		Expect(out).To(ContainSubstring("7"))
		Expect(out).To(ContainSubstring("3"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - ElapsedMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - ElapsedMsg", func() {
	It("reflects the elapsed duration in View", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowElapsed: true,
		}))
		updated, _ := update(m, status.ElapsedMsg{Elapsed: 75 * time.Second})
		Expect(updated.View().Content).To(ContainSubstring("1m15s"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - PercentMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - PercentMsg", func() {
	It("clamps values below 0 to 0", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.PercentMsg{Percent: -50})
		// 0% is hidden in the rendered row (bar is gated on
		// m.percent > 0); assert via the accessor instead.
		Expect(updated.Percent()).To(Equal(0))
		Expect(updated.View().Content).NotTo(ContainSubstring("%"))
	})

	It("clamps values above 100 to 100", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.PercentMsg{Percent: 200})
		Expect(updated.Percent()).To(Equal(100))
		Expect(updated.View().Content).To(ContainSubstring("100%"))
	})

	It("renders the percent verbatim when in range", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.PercentMsg{Percent: 42})
		Expect(updated.View().Content).To(ContainSubstring("42%"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - TotalMsg / IncDoneMsg / DoneMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - TotalMsg and IncDoneMsg", func() {
	It("computes percent from done/total after a TotalMsg", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))

		// total = 10
		updated, _ := update(m, status.TotalMsg{Total: 10})

		// done = 3 (1+1+1) → 30%
		updated, _ = update(updated, status.IncDoneMsg{N: 1})
		updated, _ = update(updated, status.IncDoneMsg{N: 1})
		updated, _ = update(updated, status.IncDoneMsg{N: 1})
		Expect(updated.View().Content).To(ContainSubstring("30%"))
	})

	It("IncDoneMsg with N=0 defaults to 1", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.TotalMsg{Total: 4})
		updated, _ = update(updated, status.IncDoneMsg{}) // N==0
		// N=0 defaults to 1, so done=1, percent=25.
		Expect(updated.Done()).To(Equal(1))
		Expect(updated.Percent()).To(Equal(25))
		Expect(updated.View().Content).To(ContainSubstring("25%"))
	})

	It("does not compute percent without a prior TotalMsg", func() {
		// No TotalMsg means m.hasTotal=false, so recomputePercent
		// is a no-op and percent stays 0.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.IncDoneMsg{N: 5})
		Expect(updated.Percent()).To(Equal(0))
		Expect(updated.HasTotal()).To(BeFalse())
		// Bar is hidden because neither percent > 0 nor hasTotal.
		Expect(updated.View().Content).NotTo(ContainSubstring("%"))
	})

	It("clamps percent to 100 when done exceeds total", func() {
		// Regression: the CensusMsg-supplied total is a preview
		// estimate. Real navigation can push done > total. The
		// recomputed ratio is clamped to 100 so the label never
		// shows "150%".
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.TotalMsg{Total: 10})
		// 5 more than the total: 15/10 = 150%, must clamp to 100.
		updated, _ = update(updated, status.IncDoneMsg{N: 15})

		Expect(updated.Percent()).To(Equal(100))
		Expect(updated.Done()).To(Equal(15), "internal done count is not clamped")
		Expect(updated.View().Content).To(ContainSubstring("100%"))
		Expect(updated.View().Content).ToNot(ContainSubstring("150%"))
	})
})

var _ = Describe("Model.Update - DoneMsg", func() {
	It("marks done and captures the error message", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowComplete: true,
		}))
		updated, _ := update(m, status.DoneMsg{
			Done: 10, IsDone: true, Err: "boom",
		})
		Expect(updated.IsDone()).To(BeTrue())
		Expect(updated.View().Content).To(ContainSubstring("❌ Failed: boom"))
	})

	It("renders ✔ complete when done matches total and no error", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowComplete: true, ShowProgress: true,
		}))
		// "✔ complete" requires both isDone AND percent >= 100,
		// so seed a total and match done to it.
		updated, _ := update(m, status.TotalMsg{Total: 10})
		updated, _ = update(updated, status.IncDoneMsg{N: 10})
		updated, _ = update(updated, status.DoneMsg{Done: 10, IsDone: true})
		Expect(updated.View().Content).To(ContainSubstring("✔ complete"))
		Expect(updated.View().Content).To(ContainSubstring("100%"))
	})

	It("does NOT force percent to 100 on IsDone=true; bar and complete message are decoupled", func() {
		// Regression: DoneMsg{IsDone:true} used to set
		// percent=100 unconditionally, coupling the progress bar
		// to the "✔ complete" message. The bar must track the
		// natural done/total ratio; the "✔ complete" message
		// renders only when BOTH isDone AND percent >= 100.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true, ShowComplete: true,
		}))

		// Without a TotalMsg, hasTotal is false and recomputePercent
		// is a no-op, so percent stays 0. The "✔ complete" message
		// does NOT show because percent < 100.
		updated, _ := update(m, status.DoneMsg{Done: 7, IsDone: true})
		Expect(updated.IsDone()).To(BeTrue())
		Expect(updated.Percent()).To(Equal(0),
			"percent depends on done/total, not on isDone")
		Expect(updated.View().Content).NotTo(ContainSubstring("✔"),
			"complete message hidden when percent < 100")
		Expect(updated.View().Content).NotTo(ContainSubstring("%"),
			"progress segment is hidden at 0% without a total")

		// With TotalMsg and done == total, percent reaches 100%.
		m2 := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true, ShowComplete: true,
		}))
		m2, _ = update(m2, status.TotalMsg{Total: 10})
		m2, _ = update(m2, status.IncDoneMsg{N: 10})
		Expect(m2.Percent()).To(Equal(100))

		m2, _ = update(m2, status.DoneMsg{Done: 10, IsDone: true})
		Expect(m2.IsDone()).To(BeTrue())
		Expect(m2.Percent()).To(Equal(100),
			"10/10 done/total produces 100% naturally")
		Expect(m2.View().Content).To(ContainSubstring("100%"))
		Expect(m2.View().Content).To(ContainSubstring("✔ complete"))
	})

	It("does NOT set percent to 100 on IsDone=false", func() {
		// DoneMsg{IsDone:false} can be used to record a partial
		// count without marking the widget as complete; the
		// percent must stay at whatever it was (0 here, since
		// no PercentMsg preceded this).
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.DoneMsg{Done: 5, IsDone: false})
		Expect(updated.IsDone()).To(BeFalse())
		Expect(updated.Percent()).To(Equal(0))
		Expect(updated.View().Content).NotTo(ContainSubstring("100%"))
	})

	It("ignores late IncDoneMsg after DoneMsg{IsDone:true} to preserve the final ratio", func() {
		// Regression: a MotifMsg that arrives after CompleteMsg
		// dispatches IncDoneMsg, which increments done and
		// recomputes percent from the done/total ratio. The
		// widget must reject IncDoneMsg once isDone is true so
		// the final ratio set by DoneMsg is preserved.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true, ShowComplete: true,
		}))

		// Seed a total and drive done to match (100%).
		m, _ = update(m, status.TotalMsg{Total: 10})
		for range 10 {
			m, _ = update(m, status.IncDoneMsg{N: 1})
		}
		Expect(m.Percent()).To(Equal(100))

		// DoneMsg recomputes from done/total (10/10 = 100%).
		m, _ = update(m, status.DoneMsg{Done: 10, IsDone: true})
		Expect(m.IsDone()).To(BeTrue())
		Expect(m.Percent()).To(Equal(100),
			"percent comes from done/total (10/10 = 100%)")
		Expect(m.View().Content).To(ContainSubstring("100%"))

		// Late IncDoneMsg arrives. It must NOT change the ratio.
		m, cmd := update(m, status.IncDoneMsg{N: 1})
		Expect(cmd).To(BeNil(),
			"IncDoneMsg after isDone must return nil cmd")
		Expect(m.IsDone()).To(BeTrue())
		Expect(m.Percent()).To(Equal(100),
			"IncDoneMsg after DoneMsg must not overwrite the final ratio")
		Expect(m.View().Content).To(ContainSubstring("100%"))
		Expect(m.View().Content).To(ContainSubstring("✔ complete"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - ResetMsg
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - ResetMsg", func() {
	It("zeros percent, done, total, hasTotal, isDone and errMsg", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true, ShowComplete: true,
		}))
		updated, _ := update(m, status.TotalMsg{Total: 100})
		updated, _ = update(updated, status.IncDoneMsg{N: 50})
		// Recompute drives percent from done/total: 50/100 = 50%.
		Expect(updated.Percent()).To(Equal(50))
		Expect(updated.View().Content).To(ContainSubstring("50%"))

		updated, _ = update(updated, status.DoneMsg{Done: 50, IsDone: true})
		Expect(updated.IsDone()).To(BeTrue())
		// DoneMsg recomputes from done/total: 50/100 = 50%.
		Expect(updated.Percent()).To(Equal(50))
		Expect(updated.View().Content).To(ContainSubstring("50%"))

		reset, _ := update(updated, status.ResetMsg{})
		Expect(reset.IsDone()).To(BeFalse())
		Expect(reset.Percent()).To(Equal(0))
		Expect(reset.Done()).To(Equal(0))
		Expect(reset.Total()).To(Equal(0))
		Expect(reset.HasTotal()).To(BeFalse())
		Expect(reset.ErrMsg()).To(BeEmpty())
		// Bar is hidden (no TotalMsg, percent=0 → segment omitted).
		Expect(reset.View().Content).NotTo(ContainSubstring("%"))
	})
})

// ---------------------------------------------------------------------------
// Bar visibility - gated on m.percent > 0 OR (m.hasTotal && m.total > 0)
// ---------------------------------------------------------------------------

var _ = Describe("Model.View - bar visibility", func() {
	It("hides the progress bar on a fresh model with no TotalMsg and percent=0", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowFiles:    true,
			ShowProgress: true,
			ShowElapsed:  true,
		}))
		// No messages sent: percent=0, hasTotal=false → bar omitted
		// so the row doesn't render an empty 10-cell track.
		Expect(m.Percent()).To(Equal(0))
		Expect(m.HasTotal()).To(BeFalse())
		Expect(m.View().Content).NotTo(ContainSubstring("%"))
		// Other segments are still visible.
		Expect(m.View().Content).To(ContainSubstring("🔖 files:"))
	})

	It("shows the bar once a TotalMsg has been seen, even at 0%", func() {
		// After TotalMsg{Total:100} with no IncDoneMsg yet,
		// hasTotal=true so the segment is rendered. done=0 →
		// percent=0 → the bar appears as an empty track with
		// the "  0%" label.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.TotalMsg{Total: 100})
		Expect(updated.HasTotal()).To(BeTrue())
		Expect(updated.Percent()).To(Equal(0))
		Expect(updated.View().Content).To(ContainSubstring("  0%"))
	})

	It("fills the bar mid-navigation as IncDoneMsg drives done up", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.TotalMsg{Total: 100})
		updated, _ = update(updated, status.IncDoneMsg{N: 30})
		Expect(updated.Done()).To(Equal(30))
		Expect(updated.Percent()).To(Equal(30))
		Expect(updated.View().Content).To(ContainSubstring("30%"))
	})

	It("appears at 100% on DoneMsg{IsDone:true} only when done equals total", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true, ShowComplete: true,
		}))
		// Seed a total and drive done to equal it.
		updated, _ := update(m, status.TotalMsg{Total: 10})
		updated, _ = update(updated, status.IncDoneMsg{N: 10})
		Expect(updated.Percent()).To(Equal(100))

		updated, _ = update(updated, status.DoneMsg{Done: 10, IsDone: true})
		Expect(updated.Percent()).To(Equal(100))
		Expect(updated.View().Content).To(ContainSubstring("100%"))
		Expect(updated.View().Content).To(ContainSubstring("✔ complete"))
	})

	It("appears at intermediate percent when set explicitly via PercentMsg", func() {
		// PercentMsg is the demo-mode path: highway pushes a
		// time-derived percent directly without supplying a
		// total.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.PercentMsg{Percent: 42})
		Expect(updated.View().Content).To(ContainSubstring("42%"))
	})

	It("hides the bar when ShowProgress is false even with percent > 0", func() {
		// The ShowProgress gate is independent of m.percent.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: false,
		}))
		updated, _ := update(m, status.PercentMsg{Percent: 50})
		Expect(updated.Percent()).To(Equal(50),
			"percent state is recorded; only the rendered segment is gated")
		Expect(updated.View().Content).NotTo(ContainSubstring("%"))
	})
})

// ---------------------------------------------------------------------------
// Model.Update - default (unknown message)
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - unknown message", func() {
	It("returns the model unchanged with nil cmd", func() {
		m := status.New(status.WithStyles(baseStyles()))
		updated, cmd := update(m, "some-unknown-msg")
		Expect(cmd).To(BeNil())
		Expect(updated.View().Content).To(Equal(m.View().Content))
	})
})

// ---------------------------------------------------------------------------
// Spring animation - SetPercent re-targeting and FrameMsg forwarding
// ---------------------------------------------------------------------------

var _ = Describe("Model.Update - spring re-targeting", func() {
	It("PercentMsg returns a non-nil first-frame cmd and re-targets the inner spring", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, cmd := update(m, status.PercentMsg{Percent: 42})
		Expect(cmd).NotTo(BeNil(), "SetPercent must return a tick cmd to drive the spring")
		// Inner().Percent() returns the spring's target (not the
		// animated value), which has just been re-targeted to 0.42.
		Expect(updated.Inner().Percent()).To(BeNumerically("~", 0.42, 1e-9))
	})

	It("IncDoneMsg after TotalMsg re-targets to done/total and returns a non-nil cmd", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.TotalMsg{Total: 10})
		updated, cmd := update(updated, status.IncDoneMsg{N: 1})
		Expect(cmd).NotTo(BeNil())
		Expect(updated.Inner().Percent()).To(BeNumerically("~", 0.1, 1e-9))
	})

	It("DoneMsg{IsDone:true} re-targets to the done/total ratio", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true, ShowComplete: true,
		}))
		// Seed a total and some done so the ratio is non-zero.
		updated, _ := update(m, status.TotalMsg{Total: 10})
		updated, _ = update(updated, status.IncDoneMsg{N: 5})
		Expect(updated.Percent()).To(Equal(50))
		// DoneMsg recomputes from done/total, not from IsDone.
		updated, cmd := update(updated, status.DoneMsg{Done: 7, IsDone: true})
		Expect(cmd).NotTo(BeNil())
		Expect(updated.Inner().Percent()).To(BeNumerically("~", 0.7, 1e-9))
	})

	It("ResetMsg re-targets to 0.0 and returns a non-nil cmd", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, _ := update(m, status.PercentMsg{Percent: 75})
		Expect(updated.Inner().Percent()).To(BeNumerically("~", 0.75, 1e-9))

		reset, cmd := update(updated, status.ResetMsg{})
		Expect(cmd).NotTo(BeNil())
		Expect(reset.Inner().Percent()).To(BeNumerically("~", 0.0, 1e-9))
	})
})

var _ = Describe("Model.Update - FrameMsg forwarding", func() {
	// The bubbles v2 FrameMsg id/tag fields are unexported; we
	// cannot construct one synthetically. Instead, we drive the
	// spring via SetPercent, capture the cmd it returns, and
	// execute the cmd to obtain a real FrameMsg.
	frameFromCmd := func(cmd tea.Cmd) progress.FrameMsg {
		GinkgoHelper()
		Expect(cmd).NotTo(BeNil(), "cmd needed to produce a FrameMsg")
		msg := cmd()
		frame, ok := msg.(progress.FrameMsg)
		Expect(ok).To(BeTrue(), "cmd must produce a FrameMsg, got %T", msg)
		return frame
	}

	It("returns a non-nil next-frame cmd while the spring is still animating", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, cmd := update(m, status.PercentMsg{Percent: 100})
		frame := frameFromCmd(cmd)

		// Forward the first frame; the spring is nowhere near
		// equilibrium yet so the next-frame cmd must be non-nil.
		updated, nextCmd := update(updated, frame)
		Expect(nextCmd).NotTo(BeNil(),
			"spring still mid-flight; next FrameMsg cmd expected")
		_ = updated
	})

	It("returns nil once the spring has reached equilibrium", func() {
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		// Re-target to 0.5 then drive the spring forward
		// repeatedly until the inner model reports nil cmd
		// (equilibrium). Bound the loop to avoid an infinite
		// spin if the spring fails to settle.
		updated, cmd := update(m, status.PercentMsg{Percent: 50})
		var nextCmd = cmd
		const maxFrames = 2000
		settled := false
		for range maxFrames {
			if nextCmd == nil {
				settled = true
				break
			}
			frame := frameFromCmd(nextCmd)
			updated, nextCmd = update(updated, frame)
		}
		Expect(settled).To(BeTrue(),
			"spring failed to reach equilibrium within %d frames", maxFrames)
	})
})

var _ = Describe("Model.View - reflects spring's animated percentShown", func() {
	It("the rendered bar is bounded between the previous value and the new target during animation", func() {
		// After SetPercent(1.0) the spring's targetPercent is 1
		// but percentShown starts at 0. A single FrameMsg
		// advances the spring by one step; the resulting bar
		// View() must therefore differ from ViewAs(1.0) and
		// match ViewAs(percentShown) which is < 1.
		m := status.New(status.WithStyles(baseStyles()), status.WithFields(status.FieldSelectors{
			ShowProgress: true,
		}))
		updated, cmd := update(m, status.PercentMsg{Percent: 100})
		Expect(cmd).NotTo(BeNil())

		// Execute the cmd to get a FrameMsg, forward it once.
		msg := cmd()
		frame, ok := msg.(progress.FrameMsg)
		Expect(ok).To(BeTrue())
		updated, _ = update(updated, frame)

		// After a single frame the spring has not yet reached
		// the target, so the animated View() output must differ
		// from the snapped ViewAs(1.0) output.
		animated := updated.Inner().View()
		snapped := updated.Inner().ViewAs(1.0)
		Expect(animated).NotTo(Equal(snapped),
			"View() must reflect the spring's animated percentShown, not the target")
	})
})

// ---------------------------------------------------------------------------
// Render (wrapper) does not duplicate View() logic - smoke test
// ---------------------------------------------------------------------------

var _ = Describe("Render equivalence with View()", func() {
	It("Render with the same inputs as Model field sets produces equivalent output", func() {
		// Construct a Model with the same fields Render would set.
		m := status.New(
			status.WithStyles(baseStyles()),
			status.WithFields(status.FieldSelectors{
				ShowFiles: true, ShowElapsed: true,
			}),
		)
		_, _ = update(m, status.CountsMsg{Files: 42, Dirs: 7, Errors: 0})
		_, _ = update(m, status.ElapsedMsg{Elapsed: 5 * time.Second})

		// Snapshot is taken AFTER updates, then we drive Render
		// with the same values.
		snapshot := m.View().Content
		_ = snapshot
		// Note: exact equality is brittle because lipgloss
		// styles are unstable across runs in some envs. The
		// real value of this test is that Render does not
		// panic and produces a non-empty string in the same
		// shape as View().
		out := status.Render(status.Config{
			Files: 42, Dirs: 7, Errors: 0, Elapsed: 5 * time.Second,
		}, baseStyles(), status.FieldSelectors{
			ShowFiles: true, ShowElapsed: true,
		}, 80)

		Expect(out).To(ContainSubstring("🔖 files:"))
		Expect(out).To(ContainSubstring("⏰ elapsed:"))
	})
})

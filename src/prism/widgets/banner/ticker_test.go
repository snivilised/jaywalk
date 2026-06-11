package banner

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/effects"
)

// White-box tests for the Ticker helper. Lives in `package banner`
// (not `package banner_test`) so it can read the private state,
// factor, and counter fields directly. The Ticker encapsulates
// mutable state that the test must inspect to verify the skip-
// counter logic - this is the same in-package test pattern used by
// status and track.

var _ = Describe("NewTicker", func() {
	It("returns nil when state is nil", func() {
		t := NewTicker(nil, 500*time.Millisecond, 50*time.Millisecond)
		Expect(t).To(BeNil())
	})

	It("uses DefaultBannerTick when bannerTick is 0", func() {
		st := effects.NewGradientState()
		t := NewTicker(st, 0, 50*time.Millisecond)
		Expect(t).NotTo(BeNil())
		// 500ms default / 50ms global = 10
		Expect(t.Factor()).To(Equal(10))
	})

	It("computes factor = bannerTick / globalTick", func() {
		st := effects.NewGradientState()
		t := NewTicker(st, 500*time.Millisecond, 50*time.Millisecond)
		Expect(t.Factor()).To(Equal(10))
	})

	It("floors the factor to 0 when bannerTick < globalTick", func() {
		st := effects.NewGradientState()
		t := NewTicker(st, 10*time.Millisecond, 50*time.Millisecond)
		Expect(t.Factor()).To(Equal(0))
	})
})

var _ = Describe("Ticker.Advance", func() {
	It("is no-op when Ticker is nil", func() {
		var t *Ticker
		// Must not panic.
		t.Advance()
	})

	It("is no-op when t.state is nil", func() {
		t := &Ticker{state: nil, factor: 1}
		// Must not panic.
		t.Advance()
		Expect(t.Counter()).To(Equal(0))
	})

	It("updates the state on every call when factor is 0", func() {
		st := newTestState()
		before := st.Offset
		t := NewTicker(st, 10*time.Millisecond, 50*time.Millisecond) // factor 0
		t.Advance()
		t.Advance()
		t.Advance()
		Expect(st.Offset).To(Equal(before + 3))
	})

	It("updates the state every factor-th call when factor > 0", func() {
		st := newTestState()
		before := st.Offset
		t := NewTicker(st, 500*time.Millisecond, 50*time.Millisecond) // factor 10
		// 9 calls: counter advances to 9, state unchanged.
		for i := 0; i < 9; i++ {
			t.Advance()
		}
		Expect(st.Offset).To(Equal(before))
		// 10th call: counter wraps, state advances by 1.
		t.Advance()
		Expect(st.Offset).To(Equal(before + 1))
	})

	It("Counter resets to 0 after a state update", func() {
		st := newTestState()
		t := NewTicker(st, 500*time.Millisecond, 50*time.Millisecond) // factor 10
		for i := 0; i < 10; i++ {
			t.Advance()
		}
		Expect(t.Counter()).To(Equal(0))
	})

	It("Ticker preserves its captured state across many Advance calls", func() {
		st := newTestState()
		t := NewTicker(st, 500*time.Millisecond, 50*time.Millisecond) // factor 10
		// 30 ticks -> 3 state advances.
		for i := 0; i < 30; i++ {
			t.Advance()
		}
		Expect(st.Offset).To(Equal(3))
		Expect(t.Counter()).To(Equal(0))
	})
})

// newTestState constructs a GradientState with a small non-empty
// step array so that Update's guard (TotalSteps > 0 && stepsArray
// non-empty) is satisfied. The actual colour values do not matter
// for these tests - only the Offset counter.
func newTestState() *effects.GradientState {
	st := effects.NewGradientState()
	st.SetSteps([]contract.Color{{}, {}, {}, {}})
	return st
}

package workshop_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/workshop"
)

type spyRenderer struct {
	calls []spyCall
}

type spyCall struct {
	Steps     []contract.Color
	Curve     enums.CurveKind
	Easing    enums.EasingKind
	AnimFrame int
}

func (s *spyRenderer) Render(
	steps []contract.Color,
	curve enums.CurveKind,
	easing enums.EasingKind,
	animFrame int,
) string {
	s.calls = append(s.calls, spyCall{
		Steps:     steps,
		Curve:     curve,
		Easing:    easing,
		AnimFrame: animFrame,
	})
	return ""
}

func (s *spyRenderer) Name() string {
	return "spy"
}

var _ = Describe("Registry", func() {
	BeforeEach(func() {
		workshop.Reset()
	})

	It("starts empty", func() {
		Expect(workshop.Visualisers()).To(BeEmpty())
	})

	It("registers a visualiser", func() {
		workshop.Register(&spyRenderer{})
		Expect(workshop.Visualisers()).To(HaveLen(1))
	})

	It("returns a copy of the registry", func() {
		workshop.Register(&spyRenderer{})
		workshop.Register(&spyRenderer{})

		v := workshop.Visualisers()
		Expect(v).To(HaveLen(2))

		workshop.Register(&spyRenderer{})
		Expect(v).To(HaveLen(2),
			"previously retrieved slice should be a snapshot",
		)
		Expect(workshop.Visualisers()).To(HaveLen(3))
	})

	It("clears on Reset", func() {
		workshop.Register(&spyRenderer{})
		Expect(workshop.Visualisers()).To(HaveLen(1))

		workshop.Reset()
		Expect(workshop.Visualisers()).To(BeEmpty())
	})

	It("RegisterVisualisers registers all four", func() {
		workshop.RegisterVisualisers()
		Expect(workshop.Visualisers()).To(HaveLen(4))

		names := make([]string, 4)
		for i, v := range workshop.Visualisers() {
			names[i] = v.Name()
		}
		Expect(names).To(ConsistOf(
			"waveform", "sweep", "bloom", "bands",
		))
	})
})

var _ = Describe("SpyRenderer", func() {
	It("records calls", func() {
		spy := &spyRenderer{}
		steps := []contract.Color{{R: 255, G: 0, B: 0}}

		result := spy.Render(
			steps,
			enums.CurveKindSine,
			enums.EasingKindEaseIn,
			5,
		)
		Expect(result).To(BeEmpty())
		Expect(spy.calls).To(HaveLen(1))
		Expect(spy.calls[0].Curve).To(
			Equal(enums.CurveKindSine),
		)
		Expect(spy.calls[0].Easing).To(
			Equal(enums.EasingKindEaseIn),
		)
		Expect(spy.calls[0].AnimFrame).To(Equal(5))
	})
})

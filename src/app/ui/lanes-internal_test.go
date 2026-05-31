package ui_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/app/ui"
	"github.com/snivilised/jaywalk/src/prism/movies"
)

var _ = Describe("BuildHighwayLanes", func() {
	BeforeEach(func() {
		movies.RegisterAll()
	})

	It("lane count is driven by now parameter, not spinner names", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"wave"},
		}
		lanes := ui.BuildHighwayLanes(cfg, 6)
		Expect(len(lanes)).To(Equal(6))
		Expect(lanes[0].SpinnerName).To(Equal("wave"))
	})

	It("single spinner name replicates the same FrameFunc across all lanes", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"wave"},
		}
		lanes := ui.BuildHighwayLanes(cfg, 6)
		Expect(len(lanes)).To(Equal(6))
		for _, lane := range lanes {
			Expect(lane.FrameFn).NotTo(BeNil())
		}
		// all lanes must share the identical FrameFunc pointer
		firstPtr := reflect.ValueOf(lanes[0].FrameFn).Pointer()
		for i := 1; i < len(lanes); i++ {
			ptr := reflect.ValueOf(lanes[i].FrameFn).Pointer()
			Expect(ptr).To(Equal(firstPtr), "lane %d has a different FrameFunc", i)
		}
	})

	It("category names expand to unique FrameFuncs across lanes", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"film-strip-set"},
		}
		lanes := ui.BuildHighwayLanes(cfg, 12)
		Expect(len(lanes)).To(Equal(12))
		funcs := make(map[uintptr]bool)
		for _, lane := range lanes {
			ptr := reflect.ValueOf(lane.FrameFn).Pointer()
			funcs[ptr] = true
		}
		Expect(len(funcs)).To(BeNumerically(">=", 10))
	})

	It("braille-set category expands to unique FrameFuncs", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"braille-set"},
		}
		lanes := ui.BuildHighwayLanes(cfg, 18)
		Expect(len(lanes)).To(Equal(18))
		funcs := make(map[uintptr]bool)
		for _, lane := range lanes {
			ptr := reflect.ValueOf(lane.FrameFn).Pointer()
			funcs[ptr] = true
		}
		Expect(len(funcs)).To(BeNumerically(">=", 17))
	})

	It("falls back to defaults when SpinnerNames is empty", func() {
		cfg := ui.HighwayConfig{}
		lanes := ui.BuildHighwayLanes(cfg, 0)
		Expect(len(lanes)).To(Equal(4))
	})

	It("all lanes have a valid FrameFunc", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"film-strip-set"},
		}
		lanes := ui.BuildHighwayLanes(cfg, 0)
		for i, lane := range lanes {
			Expect(lane.FrameFn).NotTo(BeNil(), "lane %d has nil FrameFunc", i)
		}
	})

	It("sets IntervalMs from Overrides map", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"wave", "bounce"},
			Overrides: map[string]int{
				"wave":   5000,
				"bounce": 200,
			},
		}
		lanes := ui.BuildHighwayLanes(cfg, 2)
		Expect(len(lanes)).To(Equal(2))
		Expect(lanes[0].IntervalMs).To(Equal(5000))
		Expect(lanes[1].IntervalMs).To(Equal(200))
	})

	It("sets IntervalMs to 0 for spinner names not in Overrides", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"wave"},
		}
		lanes := ui.BuildHighwayLanes(cfg, 1)
		Expect(lanes[0].IntervalMs).To(Equal(0))
	})

	It("ignores override entries with zero or negative intervals", func() {
		cfg := ui.HighwayConfig{
			SpinnerNames: []string{"wave"},
			Overrides: map[string]int{
				"wave": 0,
			},
		}
		lanes := ui.BuildHighwayLanes(cfg, 1)
		Expect(lanes[0].IntervalMs).To(Equal(0))
	})
})

package controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("workTagAllocator", func() {
	DescribeTable("deterministic sequence",
		func(tags []string, nth func(int) int, expected []string) {
			a := newWorkTagAllocator(tags, WithRandomN(nth))
			for _, exp := range expected {
				Expect(a.Allocate()).To(Equal(exp))
			}
		},
		Entry("always picks first when nth returns 0",
			[]string{"alpha", "beta", "gamma"},
			func(int) int { return 0 },
			[]string{"alpha", "beta", "gamma"},
		),
		Entry("always picks last when nth returns n-1",
			[]string{"alpha", "beta", "gamma"},
			func(n int) int { return n - 1 },
			[]string{"gamma", "beta", "alpha"},
		),
		Entry("single tag always returns the same",
			[]string{"only"},
			func(int) int { return 0 },
			[]string{"only", "only", "only"},
		),
	)

	DescribeTable("anti-repeat guarantee across cycles",
		func(tags []string) {
			a := newWorkTagAllocator(tags)
			var last string
			for range len(tags) * 3 {
				next := a.Allocate()
				Expect(next).NotTo(Equal(last))
				last = next
			}
		},
		Entry("3 tags", []string{"a", "b", "c"}),
		Entry("2 tags", []string{"x", "y"}),
		Entry("4 tags", []string{"a", "b", "c", "d"}),
	)

	It("single tag always returns the only option (repeat is inevitable)", func() {
		a := newWorkTagAllocator([]string{"only"})
		Expect(a.Allocate()).To(Equal("only"))
		Expect(a.Allocate()).To(Equal("only"))
		Expect(a.Allocate()).To(Equal("only"))
	})

	DescribeTable("cycle exhaustion with deterministic picking",
		func(tags []string, expected []string) {
			a := newWorkTagAllocator(tags, WithRandomN(func(int) int { return 0 }))
			for _, exp := range expected {
				Expect(a.Allocate()).To(Equal(exp))
			}
		},
		Entry("3 tags",
			[]string{"x", "y", "z"},
			[]string{"x", "y", "z", "x", "y", "z"},
		),
		Entry("single tag",
			[]string{"solo"},
			[]string{"solo", "solo", "solo"},
		),
		Entry("4 tags",
			[]string{"a", "b", "c", "d"},
			[]string{"a", "b", "c", "d", "a", "b", "c", "d"},
		),
		Entry("2 tags",
			[]string{"p", "q"},
			[]string{"p", "q", "p", "q"},
		),
	)

	It("panics with empty tag list", func() {
		Expect(func() {
			newWorkTagAllocator([]string{})
		}).To(PanicWith(MatchRegexp("tag list must not be empty")))
	})
})

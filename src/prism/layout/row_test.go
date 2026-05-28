package layout_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/layout"
)

var _ = Describe("Row", func() {
	Describe("basic construction", func() {
		It("renders a single content-sized segment with trailing filler", func() {
			row := layout.NewRow(20).Caps("│", "│").Content("hello")
			Expect(row.Render()).To(Equal("│hello               │"))
		})

		It("renders a fixed-width segment padded to Width", func() {
			row := layout.NewRow(20).Caps("│", "│").Fixed(10, "hi")
			Expect(row.Render()).To(Equal("│hi                  │"))
		})

		It("uses caps as border characters", func() {
			row := layout.NewRow(10).Caps("[", "]").Content("ab")
			Expect(row.Render()).To(Equal("[ab        ]"))
		})
	})

	Describe("left and right zones", func() {
		It("puts content between left and right with computed filler", func() {
			row := layout.NewRow(20).Caps("│", "│").
				Content("left").
				RightContent("right")
			Expect(row.Render()).To(Equal("│left           right│"))
		})

		It("right-aligns multiple right segments as a group", func() {
			row := layout.NewRow(20).Caps("│", "│").
				Content("A").
				RightContent("B").
				RightContent("C")
			Expect(row.Render()).To(Equal("│A                 BC│"))
		})

		It("handles gap-after on right segments before cap", func() {
			row := layout.NewRow(20).Caps("│", "│").
				Content("L").
				RightContent("R").Gap(2)
			got := row.Render()
			runes := []rune(got)
			Expect(len(runes)).To(Equal(22))
			Expect(got).To(HavePrefix("│L"))
			Expect(got).To(HaveSuffix("│"))
		})
	})

	Describe("flex segment", func() {
		It("fills remaining space between left and right zones", func() {
			row := layout.NewRow(30).Caps("│", "│").
				Content("A").
				Flex(false).
				RightContent("Z")
			row.SetFlexContent("...")
			Expect(row.Render()).To(Equal("│A...                         Z│"))
		})

		It("reports allocated width via FlexWidth", func() {
			row := layout.NewRow(20).
				Content("A").
				Flex(false).
				RightContent("B")
			Expect(row.FlexWidth()).To(Equal(18))
		})

		It("clamps flex width to minimum 1 when content overflows", func() {
			row := layout.NewRow(10).
				Content("12345").
				Flex(false).
				RightContent("67890")
			Expect(row.FlexWidth()).To(Equal(1))
		})

		It("renders content after flex left-aligned with filler before right", func() {
			row := layout.NewRow(40).Caps("│", "│").
				Content("A").
				Flex(false).Gap(2).
				Content("C").Gap(1).
				RightContent("Z")
			row.SetFlexContent("B")
			got := row.Render()
			// left-before=A(1), flex=B(1), gap=2, left-after=C(1)+gap(1)
			// right=Z(1), total=1+1+2+2+1=7, filler=40-7=33
			Expect(got).To(Equal("│AB  C " + strings.Repeat(" ", 33) + "Z│"))
		})
	})

	Describe("flex truncation", func() {
		It("truncates flex content wider than allocated width", func() {
			row := layout.NewRow(20).Caps("│", "│").
				Content("L").
				Flex(true).
				RightContent("R")
			row.SetFlexContent("verylongcontentthatshouldbetruncated")
			got := row.Render()
			Expect(got).To(ContainSubstring("…"))
			// inner width between caps = 20 runes
			inner := string([]rune(got)[1:21])
			Expect(len([]rune(inner))).To(Equal(20))
		})

		It("does not truncate when flex content fits within width", func() {
			row := layout.NewRow(20).Caps("│", "│").
				Content("L").
				Flex(false).
				RightContent("R")
			row.SetFlexContent("short")
			got := row.Render()
			Expect(got).To(ContainSubstring("short"))
			Expect(got).To(HaveSuffix("│"))
		})
	})

	Describe("gap after segments", func() {
		It("creates spacing between consecutive content segments", func() {
			row := layout.NewRow(20).Caps("│", "│").
				Content("A").Gap(3).
				Content("B")
			Expect(row.Render()).To(Equal("│A   B               │"))
		})

		It("applies gap on flex segment with right zone filler", func() {
			row := layout.NewRow(30).Caps("│", "│").
				Content("A").
				Flex(false).Gap(2).
				RightContent("Z")
			row.SetFlexContent("...")
			got := row.Render()
			// leftTotal=1, flexContent=3, flexGap=2, rightTotal=1
			// filler=30-1-3-2-1=23
			Expect(got).To(Equal("│A...                         Z│"))
		})
	})

	Describe("gap without flex", func() {
		It("inserts automatic filler between left and right", func() {
			row := layout.NewRow(10).Caps("", "").
				Content("ab").
				RightContent("xy")
			Expect(row.Render()).To(Equal("ab      xy"))
		})
	})

	Describe("empty row", func() {
		It("fills available width with spaces between caps", func() {
			row := layout.NewRow(10).Caps("[", "]")
			Expect(row.Render()).To(Equal("[          ]"))
		})
	})

	Describe("complex lane-like layout", func() {
		It("models a simplified highway lane with mixed segment types", func() {
			row := layout.NewRow(60).
				Caps("│ ", " │").
				Content("A").Gap(2).
				Fixed(10, "bbbbbbbbbb").Gap(2).
				Content("C").Gap(1).
				Fixed(8, "dddddddd").Gap(1).
				Flex(false).Gap(2).
				Content("E").Gap(1).
				RightContent("F")

			row.SetFlexContent("/path")
			got := row.Render()

			Expect(got).To(HavePrefix("│ "))
			Expect(got).To(HaveSuffix(" │"))
			// total rendered runes = caps(4) + inner(60) = 64
			Expect(len([]rune(got))).To(Equal(64))
		})
	})

	Describe("renderTo builder", func() {
		It("writes to provided strings.Builder", func() {
			row := layout.NewRow(10).Caps("<", ">").
				Content("x")
			var b strings.Builder
			row.RenderTo(&b)
			Expect(b.String()).To(Equal("<x         >"))
		})
	})

	Describe("set flex content", func() {
		It("returns the row for chaining", func() {
			row := layout.NewRow(10).Flex(false)
			chained := row.SetFlexContent("x")
			Expect(chained).To(BeIdenticalTo(row))
		})
	})
})

package contract_test

import (
	"image/color"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("ColourDistance", func() {
	Describe("NearestAnsi256", func() {
		Context("given a colour that exactly matches a palette entry", func() {
			It("returns the index of that entry", func() {
				c := &color.RGBA{R: 255, G: 0, B: 0, A: 255}
				idx := contract.NearestAnsi256(c)
				Expect(idx).To(Equal(uint8(9)))
			})

			It("returns 0 for pure black", func() {
				c := &color.RGBA{R: 0, G: 0, B: 0, A: 255}
				idx := contract.NearestAnsi256(c)
				Expect(idx).To(Equal(uint8(0)))
			})

			It("returns 15 for pure white", func() {
				c := &color.RGBA{R: 255, G: 255, B: 255, A: 255}
				idx := contract.NearestAnsi256(c)
				Expect(idx).To(Equal(uint8(15)))
			})
		})

		Context("given a colour between two palette entries", func() {
			It("returns the index of the closer entry", func() {
				c := &color.RGBA{R: 1, G: 1, B: 1, A: 255}
				idx := contract.NearestAnsi256(c)
				Expect(idx).To(Equal(uint8(0)))
			})
		})
	})

	Describe("NearestAnsi16Name", func() {
		It("exact cyan match returns cyan", func() {
			c := &color.RGBA{R: 0, G: 205, B: 205, A: 255}
			Expect(contract.NearestAnsi16Name(c)).To(Equal("cyan"))
		})

		It("exact magenta match returns magenta", func() {
			c := &color.RGBA{R: 205, G: 0, B: 205, A: 255}
			Expect(contract.NearestAnsi16Name(c)).To(Equal("magenta"))
		})

		It("colour closer to red than blue returns red", func() {
			c := &color.RGBA{R: 200, G: 0, B: 50, A: 255}
			Expect(contract.NearestAnsi16Name(c)).To(Equal("red"))
		})

		It("pure black returns black", func() {
			c := &color.RGBA{R: 0, G: 0, B: 0, A: 255}
			Expect(contract.NearestAnsi16Name(c)).To(Equal("black"))
		})

		It("pure white returns bright-white", func() {
			c := &color.RGBA{R: 255, G: 255, B: 255, A: 255}
			Expect(contract.NearestAnsi16Name(c)).To(Equal("bright-white"))
		})
	})
})

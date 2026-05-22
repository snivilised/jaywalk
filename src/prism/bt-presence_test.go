package prism_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	tea "charm.land/bubbletea/v2"
)

var _ = Describe("Bubbletea presence", func() {
	It("pins bubbletea as a direct dependency", func() {
		var _ tea.Model
		Expect(true).To(BeTrue())
	})
})

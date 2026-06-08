package highway

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

var _ = Describe("Model.FlagsRowPosition defaulting", func() {
	It("defaults to bottom when the OvertureMsg position is empty", func() {
		m := baseModel(1)
		m.FlagsRowPosition = ""
		updated, _ := update(m, OvertureMsg{OvertureMsg: contract.OvertureMsg{FlagsRowPosition: ""}})
		Expect(updated.FlagsRowPosition).To(Equal(contract.PositionBottom))
	})

	It("defaults to bottom when the OvertureMsg position is unrecognised", func() {
		m := baseModel(1)
		updated, _ := update(m, OvertureMsg{OvertureMsg: contract.OvertureMsg{FlagsRowPosition: "sideways"}})
		Expect(updated.FlagsRowPosition).To(Equal(contract.PositionBottom))
	})

	It("preserves a top position from the OvertureMsg", func() {
		m := baseModel(1)
		updated, _ := update(m, OvertureMsg{OvertureMsg: contract.OvertureMsg{FlagsRowPosition: contract.PositionTop}})
		Expect(updated.FlagsRowPosition).To(Equal(contract.PositionTop))
	})

	It("preserves an explicit bottom position from the OvertureMsg", func() {
		m := baseModel(1)
		updated, _ := update(m, OvertureMsg{OvertureMsg: contract.OvertureMsg{FlagsRowPosition: contract.PositionBottom}})
		Expect(updated.FlagsRowPosition).To(Equal(contract.PositionBottom))
	})
})

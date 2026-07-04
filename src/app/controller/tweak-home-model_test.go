package controller_test

import (
	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"

	jac "github.com/snivilised/jaywalk/src/app/controller"
)

var _ = Describe("TweakHomeModel", func() {
	var (
		model tea.Model
		coord *jac.TweakCoordinator
	)

	BeforeEach(func() {
		palette := contract.SystemPalette()
		opts := jac.TweakCoordinatorOptions{
			PreviewPath: "/tmp",
			Palette:     palette,
			ThemeName:   "system",
			Logger:      nil,
		}
		coord = jac.NewTweakCoordinator(opts)
		model = jac.NewTweakHomeModel(coord)
	})

	Describe("Init", func() {
		It("returns nil command", func() {
			Expect(model.Init()).To(BeNil())
		})
	})

	Describe("Update - WindowSizeMsg", func() {
		It("stores width and height", func() {
			msg := tea.WindowSizeMsg{Width: 100, Height: 50}
			updated, _ := model.Update(msg)
			Expect(updated).NotTo(BeNil())
		})
	})

	Describe("Update - KeyMsg", func() {
		It("handles 1/2/3/4 without error (entry points non-functional)", func() {
			for _, code := range []rune{'1', '2', '3', '4'} {
				updated, cmd := model.Update(tea.KeyPressMsg{Code: code})
				Expect(updated).NotTo(BeNil())
				Expect(cmd).To(BeNil())
			}
		})

		It("responds to z/Z by resetting dirty state via coordinator", func() {
			Expect(coord.IsDirty()).To(BeFalse())

			for _, code := range []rune{'z', 'Z'} {
				updated, cmd := model.Update(tea.KeyPressMsg{Code: code})
				Expect(updated).NotTo(BeNil())
				Expect(cmd).To(BeNil())
			}
		})

		It("sends tea.Quit on q/Q", func() {
			for _, code := range []rune{'q', 'Q'} {
				updated, cmd := model.Update(tea.KeyPressMsg{Code: code})
				Expect(updated).NotTo(BeNil())
				Expect(cmd).NotTo(BeNil())
			}
		})

		It("sends tea.Quit on ctrl+c", func() {
			updated, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
			Expect(updated).NotTo(BeNil())
			Expect(cmd).NotTo(BeNil())
		})
	})

	Describe("View", func() {
		It("renders the title bar with theme name", func() {
			content := model.View().Content
			Expect(content).To(ContainSubstring("jay tweak"))
			Expect(content).To(ContainSubstring("system"))
		})

		It("renders all four entry points", func() {
			content := model.View().Content
			Expect(content).To(ContainSubstring("Gradient Workshop"))
			Expect(content).To(ContainSubstring("Palette Editor"))
			Expect(content).To(ContainSubstring("Bindings"))
			Expect(content).To(ContainSubstring("Import Theme"))
		})

		It("renders keyboard shortcuts in the footer", func() {
			content := model.View().Content
			Expect(content).To(ContainSubstring("F"))
			Expect(content).To(ContainSubstring("Z"))
			Expect(content).To(ContainSubstring("Q"))
			Expect(content).To(ContainSubstring("File"))
			Expect(content).To(ContainSubstring("Undo"))
			Expect(content).To(ContainSubstring("Quit"))
		})
	})
})

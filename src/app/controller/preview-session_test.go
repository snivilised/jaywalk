package controller_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/agenor/core"
	jac "github.com/snivilised/jaywalk/src/app/controller"
)

var _ = Describe("PreviewSession", func() {
	var session *jac.PreviewSession

	BeforeEach(func() {
		session = jac.NewPreviewSession()
	})

	Describe("NewPreviewSession", func() {
		It("starts with IsComplete == false", func() {
			Expect(session.IsComplete()).To(BeFalse())
		})
	})

	Describe("MarkComplete", func() {
		It("sets IsComplete to true", func() {
			session.MarkComplete()
			Expect(session.IsComplete()).To(BeTrue())
		})

		It("can be called multiple times (idempotent)", func() {
			session.MarkComplete()
			session.MarkComplete()
			Expect(session.IsComplete()).To(BeTrue())
		})
	})

	Describe("StartedAt", func() {
		It("returns the zero time", func() {
			Expect(session.StartedAt()).To(Equal(time.Time{}))
		})
	})

	Describe("Elapsed", func() {
		It("returns 0", func() {
			Expect(session.Elapsed()).To(Equal(time.Duration(0)))
		})
	})

	Describe("core.Session implementation", func() {
		It("satisfies the core.Session interface", func() {
			var s core.Session = session
			Expect(s).NotTo(BeNil())
		})

		It("can be assigned to a core.Session variable", func() {
			var s core.Session = session
			Expect(s.IsComplete()).To(BeFalse())
			session.MarkComplete()
			Expect(s.IsComplete()).To(BeTrue())
		})
	})
})

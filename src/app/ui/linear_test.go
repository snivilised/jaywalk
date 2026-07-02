package ui_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/app/ui"
	"github.com/snivilised/jaywalk/src/prism/contract"

	"reflect"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Spy renderer
// ---------------------------------------------------------------------------

// spyRenderer captures calls made to it by the linear presenter so
// that tests can assert on the translated prism types without depending
// on terminal output or lipgloss rendering.
type spyRenderer struct {
	overture contract.Overture
	motifs   []contract.Motif
	summary  contract.Summary

	beginCalled bool
	endCalled   bool
}

func (s *spyRenderer) Begin(o contract.Overture) {
	s.overture = o
	s.beginCalled = true
}

func (s *spyRenderer) Show(m contract.Motif) {
	s.motifs = append(s.motifs, m)
}

func (s *spyRenderer) End(su contract.Summary) {
	s.summary = su
	s.endCalled = true
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// newLinearWithSpy constructs a ui.linearPresenter backed by the given spy and
// returns it as a report.Presenter. This uses the exported
// ui.NewLinearWithRenderer constructor added specifically for testing
// so that spies can be injected without going through the registry.
func newLinearWithSpy(spy contract.Renderer) report.Presenter {
	return ui.NewLinearWithRenderer(spy)
}

// stubNode builds a minimal *core.Node for use in event structs.
func stubNode(path, name string, isDir bool, depth core.TraversalDepth) *core.Node {
	n := &core.Node{
		Path: path,
		Extension: core.Extension{
			Name:  name,
			Depth: depth,
		},
	}

	// Use reflection to set the unexported 'dir' field.
	v := reflect.ValueOf(n).Elem()
	f := v.FieldByName("dir")
	//nolint:gosec // unsafe is used here to set a private field for testing purposes only
	f = reflect.NewAt(f.Type(), unsafe.Pointer(f.UnsafeAddr())).Elem()
	f.SetBool(isDir)

	return n
}

// ---------------------------------------------------------------------------
// Specs
// ---------------------------------------------------------------------------

var _ = Describe("linear", Ordered, func() {
	var (
		spy       *spyRenderer
		presenter report.Presenter
	)

	BeforeEach(func() {
		spy = &spyRenderer{}
		presenter = newLinearWithSpy(spy)
	})

	// ------------------------------------------------------------------
	// OnBegin
	// ------------------------------------------------------------------

	Describe("OnBegin", func() {
		Context("for a prime traversal", func() {
			It("passes PrimeNavigation kind and root to renderer.Begin", func() {
				now := core.Now()

				presenter.OnBegin(&report.BeginEvent{
					Root:       "/home/user/docs",
					Caption:    "files and folders",
					StartedAt:  now,
					IsPrime:    true,
					DateFormat: "Mon, 02 Jan 2006 15:04:05 MST",
				})

				Expect(spy.beginCalled).To(BeTrue())
				Expect(spy.overture.Root).To(Equal("/home/user/docs"))
				Expect(spy.overture.Caption).To(Equal("files and folders"))
				Expect(spy.overture.StartedAt).To(Equal(now))
				Expect(spy.overture.Kind).To(Equal(enums.NavigationKindPrime))
				Expect(spy.overture.ResumeFrom).To(BeEmpty())
			})
		})

		Context("for a resume traversal", func() {
			It("passes ResumeNavigation kind and resume path to renderer.Begin", func() {
				presenter.OnBegin(&report.BeginEvent{
					Root:       "/home/user/docs",
					Caption:    "files only",
					StartedAt:  core.Now(),
					IsPrime:    false,
					ResumeFrom: "/home/user/docs/subdir",
					DateFormat: "Mon, 02 Jan 2006 15:04:05 MST",
				})

				Expect(spy.overture.Kind).To(Equal(enums.NavigationKindResume))
				Expect(spy.overture.ResumeFrom).To(Equal("/home/user/docs/subdir"))
			})
		})
	})

	// ------------------------------------------------------------------
	// OnNodeEvent
	// ------------------------------------------------------------------

	Describe("OnNodeEvent", func() {
		DescribeTable("translates node fields into Motif correctly",
			func(path, name string, isDir bool, depth core.TraversalDepth) {
				presenter.OnBegin(&report.BeginEvent{
					Root:    path,
					IsPrime: true,
				})

				node := stubNode(path, name, isDir, depth)
				presenter.OnNodeEvent(&report.NeutralEvent{
					DisplayEvent: report.DisplayEvent{Node: node},
				})

				Expect(spy.motifs).To(HaveLen(1))
				m := spy.motifs[0]
				Expect(m.Path).To(Equal(path))
				Expect(m.Name).To(Equal(name))
				Expect(m.Depth).To(Equal(depth))
				Expect(m.ActionName).To(BeEmpty())
				Expect(m.PipelineName).To(BeEmpty())
				Expect(m.Skipped).To(BeFalse())
				Expect(m.Err).To(BeNil())
			},
			Entry("root directory at depth 0",
				"/home/user/docs", "docs", true, core.TraversalDepth(0),
			),
			Entry("file at depth 1",
				"/home/user/docs/report.pdf", "report.pdf", false, core.TraversalDepth(1),
			),
			Entry("nested directory at depth 3",
				"/a/b/c/d", "d", true, core.TraversalDepth(3),
			),
		)
	})

	// ------------------------------------------------------------------
	// OnActionEvent
	// ------------------------------------------------------------------

	Describe("OnActionEvent", func() {
		Context("when the action succeeds", func() {
			It("sets ActionName and leaves Err nil", func() {
				node := stubNode("/docs/file.mp4", "file.mp4", false, 1)
				presenter.OnActionEvent(&report.ActionEvent{
					DisplayEvent: report.DisplayEvent{
						Node: node,
						Name: "encode",
					},
					Err: nil,
				})

				Expect(spy.motifs).To(HaveLen(1))
				m := spy.motifs[0]
				Expect(m.ActionName).To(Equal("encode"))
				Expect(m.Err).To(BeNil())
			})
		})

		Context("when the action fails", func() {
			It("sets ActionName and Err on the Motif", func() {
				actionErr := errors.New("ffmpeg: codec not found")
				node := stubNode("/docs/file.mp4", "file.mp4", false, 1)

				presenter.OnActionEvent(&report.ActionEvent{
					DisplayEvent: report.DisplayEvent{
						Node: node,
						Name: "encode",
					},
					Err: actionErr,
				})

				Expect(spy.motifs).To(HaveLen(1))
				m := spy.motifs[0]
				Expect(m.ActionName).To(Equal("encode"))
				Expect(m.Err).To(MatchError(actionErr))
			})
		})
	})

	// ------------------------------------------------------------------
	// OnPipelineEvent
	// ------------------------------------------------------------------

	Describe("OnPipelineEvent", func() {
		Context("when the pipeline starts (header)", func() {
			It("sets PipelineName and IsPipelineHeader on the Motif", func() {
				node := stubNode("/docs/file.mp4", "file.mp4", false, 1)

				presenter.OnPipelineEvent(&report.PipelineEvent{
					DisplayEvent: report.DisplayEvent{
						Node:             node,
						Name:             "encode-and-upload",
						IsPipelineHeader: true,
					},
				})

				Expect(spy.motifs).To(HaveLen(1))
				m := spy.motifs[0]
				Expect(m.PipelineName).To(Equal("encode-and-upload"))
				Expect(m.IsPipelineHeader).To(BeTrue())
				Expect(m.IsPipelineStep).To(BeFalse())
			})
		})

		Context("when the pipeline succeeds", func() {
			It("sets PipelineName and leaves Err nil", func() {
				node := stubNode("/docs/file.mp4", "file.mp4", false, 1)

				presenter.OnPipelineEvent(&report.PipelineEvent{
					DisplayEvent: report.DisplayEvent{
						Node: node,
						Name: "encode-and-upload",
					},
				})

				Expect(spy.motifs).To(HaveLen(1))
				m := spy.motifs[0]
				Expect(m.PipelineName).To(Equal("encode-and-upload"))
				Expect(m.Err).To(BeNil())
			})
		})
	})

	Describe("OnActionEvent (Pipeline Step)", func() {
		It("sets IsPipelineStep, IsLastStep and increments VisualDepth", func() {
			node := stubNode("/docs/file.mp4", "file.mp4", false, 1) // node VisualDepth is 2

			presenter.OnActionEvent(&report.ActionEvent{
				DisplayEvent: report.DisplayEvent{
					Node:           node,
					Name:           "step1",
					IsPipelineStep: true,
					IsLastStep:     false,
				},
			})

			Expect(spy.motifs).To(HaveLen(1))
			m := spy.motifs[0]
			Expect(m.ActionName).To(Equal("step1"))
			Expect(m.IsPipelineStep).To(BeTrue())
			Expect(m.IsLastStep).To(BeFalse())
			Expect(m.VisualDepth).To(Equal(node.VisualDepth() + 1))
		})

		It("handles the last step correctly", func() {
			node := stubNode("/docs/file.mp4", "file.mp4", false, 1)

			presenter.OnActionEvent(&report.ActionEvent{
				DisplayEvent: report.DisplayEvent{
					Node:           node,
					Name:           "step2",
					IsPipelineStep: true,
					IsLastStep:     true,
				},
			})

			Expect(spy.motifs).To(HaveLen(1))
			m := spy.motifs[0]
			Expect(m.IsLastStep).To(BeTrue())
		})
	})

	// ------------------------------------------------------------------
	// OnSkipEvent
	// ------------------------------------------------------------------

	Describe("OnSkipEvent", func() {
		It("sets Skipped, Placeholder and ResolvedPath on the Motif", func() {
			node := stubNode("/docs/file.mp4", "file.mp4", false, 1)

			presenter.OnSkipEvent(&report.SkipEvent{
				DisplayEvent: report.DisplayEvent{
					Node: node,
					Name: "encode",
				},
				Placeholder:  "{{.path}}",
				ResolvedPath: "/",
			})

			Expect(spy.motifs).To(HaveLen(1))
			m := spy.motifs[0]
			Expect(m.Skipped).To(BeTrue())
			Expect(m.Placeholder).To(Equal("{{.path}}"))
			Expect(m.ResolvedPath).To(Equal("/"))
			Expect(m.ActionName).To(Equal("encode"))
		})

		It("handles pipeline step skips correctly", func() {
			node := stubNode("/docs/file.mp4", "file.mp4", false, 1)

			presenter.OnSkipEvent(&report.SkipEvent{
				DisplayEvent: report.DisplayEvent{
					Node:           node,
					Name:           "step1",
					IsPipelineStep: true,
					IsLastStep:     true,
				},
				Placeholder:  "{{.path}}",
				ResolvedPath: "/",
			})

			Expect(spy.motifs).To(HaveLen(1))
			m := spy.motifs[0]
			Expect(m.IsPipelineStep).To(BeTrue())
			Expect(m.IsLastStep).To(BeTrue())
			Expect(m.VisualDepth).To(Equal(node.VisualDepth() + 1))
		})
	})

	// ------------------------------------------------------------------
	// OnComplete
	// ------------------------------------------------------------------

	Describe("OnComplete", func() {
		Context("for a successful prime traversal", func() {
			It("passes counts, elapsed, and PrimeNavigation kind to renderer.End", func() {
				presenter.OnBegin(&report.BeginEvent{
					Root:    "/docs",
					IsPrime: true,
				})

				traversal := &report.Traversal{
					FilesVisited: 42,
					DirsVisited:  7,
					Elapsed:      3 * time.Second,
				}

				presenter.OnComplete(traversal)

				Expect(spy.endCalled).To(BeTrue())
				s := spy.summary
				Expect(s.FilesVisited).To(Equal(core.MetricValue(42)))
				Expect(s.DirsVisited).To(Equal(core.MetricValue(7)))
				Expect(s.Elapsed).To(Equal(3 * time.Second))
				Expect(s.Errors).To(BeEmpty())
				Expect(s.Kind).To(Equal(enums.NavigationKindPrime))
			})
		})

		Context("when the traversal ended with an error", func() {
			It("includes the error in Summary.Errors", func() {
				traversalErr := errors.New("permission denied")

				presenter.OnBegin(&report.BeginEvent{
					Root:    "/docs",
					IsPrime: true,
				})

				traversal := &report.Traversal{
					FilesVisited: 5,
					Err:          traversalErr,
				}

				presenter.OnComplete(traversal)

				s := spy.summary
				Expect(s.Errors).To(HaveLen(1))
				Expect(s.Errors[0]).To(MatchError(traversalErr))
			})
		})

		Context("for a resume traversal", func() {
			It("passes ResumeNavigation kind to renderer.End", func() {
				presenter.OnBegin(&report.BeginEvent{
					Root:    "/docs",
					IsPrime: false,
				})

				presenter.OnComplete(&report.Traversal{})

				Expect(spy.summary.Kind).To(Equal(enums.NavigationKindResume))
			})
		})
	})

	// ------------------------------------------------------------------
	// SubscribeFiles mode
	// ------------------------------------------------------------------

	Describe("SubscribeFiles mode", func() {
		It("injects all missing ancestors top-down", func() {
			// Hierarchy:
			// /app (depth 0)
			//   a/ (depth 1)
			//     b/ (depth 2)
			//       file.txt (depth 3)

			root := stubNode("/app", "app", true, 0)
			a := stubNode("/app/a", "a", true, 1)
			a.Parent = root
			b := stubNode("/app/a/b", "b", true, 2)
			b.Parent = a
			file := stubNode("/app/a/b/file.txt", "file.txt", false, 3)
			file.Parent = b

			presenter.OnBegin(&report.BeginEvent{
				Root:         "/app",
				Subscription: enums.SubscribeFiles,
			})

			presenter.OnActionEvent(&report.ActionEvent{
				DisplayEvent: report.DisplayEvent{
					Node: file,
					Name: "run",
				},
			})

			// Expect:
			// 1. Motif for app/ (injected)
			// 2. Motif for a/ (injected)
			// 3. Motif for b/ (injected)
			// 4. Motif for file.txt (actual)
			Expect(spy.motifs).To(HaveLen(4))
			Expect(spy.motifs[0].Name).To(Equal("app"))
			Expect(spy.motifs[1].Name).To(Equal("a"))
			Expect(spy.motifs[2].Name).To(Equal("b"))
			Expect(spy.motifs[3].Name).To(Equal("file.txt"))

			// Verify visual depths
			Expect(spy.motifs[0].VisualDepth).To(Equal(core.TraversalDepth(0)))
			Expect(spy.motifs[1].VisualDepth).To(Equal(core.TraversalDepth(1)))
			Expect(spy.motifs[2].VisualDepth).To(Equal(core.TraversalDepth(2)))
			Expect(spy.motifs[3].VisualDepth).To(Equal(core.TraversalDepth(4))) // file is depth+1
		})

		It("uses peerInfo to set IsLast on injected parents", func() {
			root := stubNode("/app", "app", true, 0)
			a := stubNode("/app/a", "a", true, 1)
			a.Parent = root
			file := stubNode("/app/a/file.txt", "file.txt", false, 2)
			file.Parent = a

			peerInfoMap := map[string]*core.PeerInfo{
				"/app":   {IsLast: true},
				"/app/a": {IsLast: false},
			}

			presenter.OnBegin(&report.BeginEvent{
				Root:         "/app",
				Subscription: enums.SubscribeFiles,
			})
			pa, ok := presenter.(report.PeerAware)
			Expect(ok).To(BeTrue())
			pa.OnPeerInfoBegin(1, 2, peerInfoMap)

			presenter.OnActionEvent(&report.ActionEvent{
				DisplayEvent: report.DisplayEvent{
					Node: file,
				},
			})

			Expect(spy.motifs[0].Name).To(Equal("app"))
			Expect(spy.motifs[0].IsLast).To(BeTrue())
			Expect(spy.motifs[1].Name).To(Equal("a"))
			Expect(spy.motifs[1].IsLast).To(BeFalse())
		})
	})
})

//go:build !race
// +build !race

package ui

import (
	"reflect"
	"time"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// killProgram safely shuts down a bubbletea program if it exists.
// This prevents terminal corruption from leaked programs.
func killProgram(p report.Presenter) {
	pp, ok := p.(*portholePresenter)
	if !ok || pp.program == nil {
		return
	}

	pp.program.Quit()

	// Wait for done channel with timeout to avoid hanging
	if pp.done != nil {
		select {
		case <-pp.done:
		case <-time.After(100 * time.Millisecond):
		}
	}
}

// stubNode builds a minimal *core.Node for use in event structs.
func stubNode(path, name string, isDir bool, depth core.TraversalDepth) *core.Node { //nolint:unparam // path: "/test/path"
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

var _ = Describe("Porthole Presenter", func() {
	var (
		presenter *portholePresenter
		palette   contract.Palette
	)

	BeforeEach(func() {
		palette = contract.SystemPalette()
	})

	Describe("Presenter creation", func() {
		It("creates a presenter with default config", func() {
			cfg := PortholeConfig{
				WorkerPool:        "🚀",
				JobPool:           "✅",
				Separator:         " ",
				AnimationGradient: "default",
			}

			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())
			Expect(p).NotTo(BeNil())
		})

		It("creates a presenter with empty config", func() {
			cfg := PortholeConfig{}

			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())
			Expect(p).NotTo(BeNil())
		})
	})

	Describe("OnTraversalOptions", func() {
		It("sets noW from options", func() {
			presenter = &portholePresenter{}
			opts := pref.Options{}
			opts.Concurrency.NoW = 4

			presenter.OnTraversalOptions(&opts)
			Expect(presenter.noW).To(Equal(uint(4)))
		})
	})

	Describe("OnBegin", func() {
		It("initializes the bubbletea program", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			e := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
				Cancel:       nil,
			}

			p.OnBegin(e)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
			defer killProgram(p)
		})

		It("sends OvertureMsg with correct fields", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			e := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeFiles,
				StartedAt:    core.Now(),
				Cancel:       nil,
			}

			p.OnBegin(e)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
			defer killProgram(p)
		})

		It("sets banner info with correct defaults", func() {
			cfg := PortholeConfig{
				Banner: BannerConfig{
					Position: contract.PositionTop,
				},
			}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			bannerInfo := p.(*portholePresenter).buildBannerInfo() //nolint:errcheck // ok
			Expect(bannerInfo.Position).To(Equal(contract.PositionTop))
		})

		It("populates bannerInfo on the presenter during OnBegin", func() {
			cfg := PortholeConfig{
				Banner: BannerConfig{
					Position: contract.PositionTop,
				},
			}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			pp := p.(*portholePresenter) //nolint:errcheck // ok
			Expect(pp.bannerInfo.Position).To(Equal(""),
				"bannerInfo should be zero before OnBegin")

			e := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
			}
			pp.OnBegin(e)
			defer killProgram(p)

			Expect(pp.bannerInfo.Position).To(Equal(contract.PositionTop),
				"OnBegin must call buildBannerInfo and store result")
		})
	})

	Describe("OnNodeEvent", func() {
		It("sends ContentLineMsg for neutral events", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			beginEvent := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
			}
			p.OnBegin(beginEvent)
			defer killProgram(p)

			e := &report.NeutralEvent{
				DisplayEvent: report.DisplayEvent{
					Node: stubNode("/test/path", "file.txt", false, 1),
				},
			}

			p.OnNodeEvent(e)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
		})
	})

	Describe("OnComplete", func() {
		It("sends CompleteMsg with traversal stats", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			beginEvent := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
			}
			p.OnBegin(beginEvent)

			// Start a goroutine to quit the program after a short delay
			// This prevents the test from hanging on <-p.done
			go func() {
				time.Sleep(50 * time.Millisecond)
				pp := p.(*portholePresenter) //nolint:errcheck // ok
				if pp.program != nil {
					pp.program.Quit()
				}
			}()

			traversal := &report.Traversal{
				FilesVisited: 10,
				DirsVisited:  5,
				Elapsed:      time.Second * 2,
				Err:          nil,
			}

			p.OnComplete(traversal)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
		})
	})

	Describe("NeedsPeerInfo", func() {
		It("returns true for porthole presenter", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			Expect(p.(report.PeerAware).NeedsPeerInfo()).To(BeTrue()) //nolint:errcheck // ok
		})
	})

	Describe("OnPeerInfoBegin", func() {
		It("sets total files and dirs from peer info", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			info := map[string]*core.PeerInfo{
				"peer1": {},
			}

			p.(report.PeerAware).OnPeerInfoBegin(100, 50, info)            //nolint:errcheck // ok
			Expect(p.(*portholePresenter).totalFiles).To(Equal(uint(100))) //nolint:errcheck // ok
			Expect(p.(*portholePresenter).totalDirs).To(Equal(uint(50)))   //nolint:errcheck // ok
		})
	})

	Describe("OnPeerInfoEnd", func() {
		It("does nothing on peer info end", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			p.(report.PeerAware).OnPeerInfoEnd() //nolint:errcheck // ok
			// No panic expected
		})
	})

	Describe("OnActionEvent", func() {
		It("sends ContentLineMsg for action events", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			beginEvent := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
			}
			p.OnBegin(beginEvent)
			defer killProgram(p)

			e := &report.ActionEvent{
				DisplayEvent: report.DisplayEvent{
					Node: stubNode("/test/path", "file.txt", false, 1),
					Name: "my-action",
				},
				CommandOutput:   "",
				ExecutionString: "ls -la",
				DryRun:          false,
				Err:             nil,
			}

			p.OnActionEvent(e)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
		})
	})

	Describe("OnPipelineEvent", func() {
		It("sends ContentLineMsg for pipeline events", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			beginEvent := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
			}
			p.OnBegin(beginEvent)
			defer killProgram(p)

			e := &report.PipelineEvent{
				DisplayEvent: report.DisplayEvent{
					Node: stubNode("/test/path", "file.txt", false, 1),
					Name: "my-pipeline",
				},
				CommandOutput:   "",
				ExecutionString: "ls -la",
				DryRun:          false,
				Err:             nil,
			}

			p.OnPipelineEvent(e)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
		})
	})

	Describe("OnSkipEvent", func() {
		It("sends ContentLineMsg for skip events", func() {
			cfg := PortholeConfig{}
			p, err := newPortholePresenter(palette, cfg)
			Expect(err).To(BeNil())

			beginEvent := &report.BeginEvent{
				Root:         "/test/root",
				Caption:      "Test Caption",
				Subscription: enums.SubscribeUniversal,
				StartedAt:    core.Now(),
			}
			p.OnBegin(beginEvent)
			defer killProgram(p)

			e := &report.SkipEvent{
				DisplayEvent: report.DisplayEvent{
					Node: stubNode("/test/path", "file.txt", false, 1),
					Name: "skip-action",
				},
				Placeholder:  "{{.path}}",
				ResolvedPath: "/",
			}

			p.OnSkipEvent(e)
			Expect(p.(*portholePresenter).program).NotTo(BeNil()) //nolint:errcheck // ok
		})
	})
})

package porthole

import (
	"io"
	"time"

	bp "charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/movies"
)

func testTheme() contract.Theme {
	t, err := contract.NewTheme(contract.SystemPalette(), io.Discard)
	if err != nil {
		panic(err)
	}
	return t
}

func baseModel() Model {
	return NewModel(contract.NewModelParams{
		RootPath:  "/root",
		MaxDepth:  5,
		Theme:     testTheme(),
		NoRecurse: false,
	})
}

func update(m Model, msg tea.Msg) (Model, tea.Cmd) {
	r, cmd := m.Update(msg)
	return *r.(*Model), cmd //nolint:errcheck // ok
}

func applyOverture(m Model) Model {
	updated, _ := update(m, OvertureMsg{
		OvertureMsg: contract.OvertureMsg{
			Root:    "/root",
			Caption: "files and folders",
		},
	})
	return updated
}

func contentLine(path, name string) ContentLineMsg {
	return ContentLineMsg{
		Line: path,
		Params: RenderParams{
			NodeParams: contract.NodeParams{
				Path:  path,
				Name:  name,
				IsDir: false,
				Depth: 1,
			},
		},
	}
}

var _ = Describe("Porthole Model", Ordered, func() {
	BeforeAll(func() {
		movies.RegisterAll()
	})

	Describe("Progress bar", func() {
		It("shows progress percentage after receiving CensusMsg and ContentLineMsg", func() {
			m := baseModel()
			m = applyOverture(m)

			m, _ = update(m, CensusMsg{TotalFiles: 10, TotalDirs: 0})
			Expect(m.status.HasTotal()).To(BeTrue())
			Expect(m.status.Total()).To(Equal(10))

			m, _ = update(m, contentLine("/root/f1.txt", "f1.txt"))
			Expect(m.status.Done()).To(Equal(1))
			Expect(m.status.Percent()).To(Equal(10))
			Expect(m.status.View().Content).To(ContainSubstring("10%"))

			m, _ = update(m, contentLine("/root/f2.txt", "f2.txt"))
			Expect(m.status.Done()).To(Equal(2))
			Expect(m.status.Percent()).To(Equal(20))
			Expect(m.status.View().Content).To(ContainSubstring("20%"))
		})

		It("correctly computes progress from files and dirs combined", func() {
			m := baseModel()
			m = applyOverture(m)

			m, _ = update(m, CensusMsg{TotalFiles: 3, TotalDirs: 2})
			Expect(m.status.Total()).To(Equal(5))

			for _, p := range []string{"/r/a.txt", "/r/b.txt", "/r/c.txt"} {
				m, _ = update(m, contentLine(p, "f"))
			}
			Expect(m.status.Done()).To(Equal(3))
			Expect(m.status.Percent()).To(Equal(60))

			m, _ = update(m, ContentLineMsg{
				Line: "/r/sub1",
				Params: RenderParams{
					NodeParams: contract.NodeParams{
						Path:  "/r/sub1",
						Name:  "sub1",
						IsDir: true,
						Depth: 1,
					},
				},
			})
			Expect(m.status.Done()).To(Equal(4))
			Expect(m.status.Percent()).To(Equal(80))

			m, _ = update(m, ContentLineMsg{
				Line: "/r/sub2",
				Params: RenderParams{
					NodeParams: contract.NodeParams{
						Path:  "/r/sub2",
						Name:  "sub2",
						IsDir: true,
						Depth: 1,
					},
				},
			})
			Expect(m.status.Done()).To(Equal(5))
			Expect(m.status.Percent()).To(Equal(100))
		})

		It("forwards bp.FrameMsg to the status widget for spring animation", func() {
			m := baseModel()
			m = applyOverture(m)

			updated, cmd := update(m, CensusMsg{TotalFiles: 10})
			Expect(cmd).NotTo(BeNil(), "CensusMsg must propagate the spring cmd")

			updated, cmd = update(updated, contentLine("/root/f.txt", "f.txt"))
			Expect(cmd).NotTo(BeNil())

			msg := cmd()
			frame, ok := msg.(bp.FrameMsg)
			Expect(ok).To(BeTrue(), "spring cmd must yield a bp.FrameMsg, got %T", msg)

			_, nextCmd := update(updated, frame)
			Expect(nextCmd).NotTo(BeNil(),
				"FrameMsg must reach the spring and produce a next-frame cmd")
		})

		It("reaches 100% on CompleteMsg regardless of CensusMsg total", func() {
			m := baseModel()
			m = applyOverture(m)

			m, _ = update(m, CensusMsg{TotalFiles: 100})

			m, cmd := update(m, CompleteMsg{Files: 80, Dirs: 5, Elapsed: 2 * time.Second})
			_ = cmd

			Expect(m.status.IsDone()).To(BeTrue())
			Expect(m.status.Percent()).To(Equal(100))
			Expect(m.status.Files()).To(Equal(80))
			Expect(m.status.Dirs()).To(Equal(5))
			Expect(m.status.View().Content).To(ContainSubstring("100%"))
			Expect(m.status.View().Content).To(ContainSubstring("✔ complete"))
		})
	})
})

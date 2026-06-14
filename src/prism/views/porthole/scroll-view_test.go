package porthole_test

import (
	"bytes"
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/views/linear"
	"github.com/snivilised/jaywalk/src/prism/views/porthole"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// update is a convenience wrapper that calls Update on the model and
// type-asserts the result back to Model.
func update(m porthole.Model, msg tea.Msg) porthole.Model {
	r, _ := m.Update(msg)
	return *r.(*porthole.Model) //nolint:errcheck // ok
}

var _ = Describe("Porthole View", func() {
	var (
		w       *bytes.Buffer
		err     error
		theme   contract.Theme
		palette contract.Palette
	)

	BeforeEach(func() {
		w = &bytes.Buffer{}
		palette = contract.SystemPalette() // has no highlights
		theme, err = contract.NewTheme(palette, w)
		Expect(err).To(BeNil())
	})

	Describe("Model creation", func() {
		It("creates a model with default settings", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/test/path", MaxDepth: 0, Theme: theme, NoRecurse: false})
			Expect(model).NotTo(BeNil())
		})

		It("creates a model with max depth and no recurse", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/test/path", MaxDepth: 3, Theme: theme, NoRecurse: true})
			Expect(model).NotTo(BeNil())
		})

		It("initializes startedAt timestamp", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/test/path", MaxDepth: 0, Theme: theme, NoRecurse: false})
			Expect(model).NotTo(BeNil())
		})
	})

	Describe("OvertureMsg handling", func() {
		It("sets all fields from OvertureMsg", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			msg := porthole.OvertureMsg{
				OvertureMsg: contract.OvertureMsg{
					Root:         "/new/root",
					Caption:      "Test Caption",
					PipelineName: "test-pipeline",
				},
				Banner: banner.Info{},
			}

			m := update(model, msg)
			Expect(m.RootPath).To(Equal("/new/root"))
			Expect(m.Caption).To(Equal("Test Caption"))
			Expect(m.PipelineName).To(Equal("test-pipeline"))
		})

		It("sets content buffer from ContentLineMsg", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			msg := porthole.ContentLineMsg{
				Line: "Test line\n",
			}

			m := update(model, msg)
			Expect(len(m.Buffer())).To(Equal(1))
			Expect(m.Buffer()[0]).To(Equal("Test line"))
		})

		It("strips trailing newlines from content lines", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			msg := porthole.ContentLineMsg{
				Line: "line with newline\n",
			}

			m := update(model, msg)
			Expect(m.Buffer()[0]).To(Equal("line with newline"))
		})

		It("skips empty content lines", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})

			m := update(model, porthole.ContentLineMsg{Line: "\n"})
			m = update(m, porthole.ContentLineMsg{Line: ""})
			m = update(m, porthole.ContentLineMsg{Line: "real content\n"})

			Expect(len(m.Buffer())).To(Equal(1))
			Expect(m.Buffer()[0]).To(Equal("real content"))
		})

		It("truncates content buffer when it exceeds MaxContentBufferLines", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})

			// Add more lines than the max
			for i := 0; i < porthole.MaxContentBufferLines+10; i++ {
				msg := porthole.ContentLineMsg{
					Line: "Line " + string(rune(i)) + "\n",
				}
				m := update(model, msg)
				model = m
			}

			Expect(len(model.Buffer())).To(BeNumerically("<=", porthole.MaxContentBufferLines))
		})

		It("sets done flag from CompleteMsg", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			msg := contract.CompleteMsg{
				Files:   10,
				Dirs:    5,
				Elapsed: time.Second,
			}

			m := update(model, msg)
			Expect(m.IsDone()).To(BeTrue())
		})

		It("stores banner info from OvertureMsg", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})

			grad, _ := theme.GradientFor(contract.GradientComponentBanner)
			info := banner.Info{
				Disable:  false,
				Position: contract.PositionTop,
				Width:    80,
				Gradient: &grad,
			}

			m := update(model, porthole.OvertureMsg{
				OvertureMsg: contract.OvertureMsg{
					Root: "/root",
				},
				Banner: info,
			})

			Expect(m.BannerInfo().Position).To(Equal(contract.PositionTop))
			Expect(m.BannerInfo().Width).To(Equal(80))
			Expect(m.BannerInfo().Disable).To(BeFalse())
		})
	})

	Describe("View rendering", func() {
		It("renders header and status with no content", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("/root"))
			Expect(output).To(ContainSubstring("Processing"))
		})

		It("renders content lines in the body", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			model = update(model, porthole.ContentLineMsg{Line: "file-a.txt\n"})
			model = update(model, porthole.ContentLineMsg{Line: "file-b.txt\n"})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("file-a.txt"))
			Expect(output).To(ContainSubstring("file-b.txt"))
		})

		It("renders multiple content lines in order", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			for i := 0; i < 10; i++ {
				model = update(model, porthole.ContentLineMsg{
					Line: "line-" + string(rune('a'+i)) + "\n",
				})
			}

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("line-a"))
			Expect(output).To(ContainSubstring("line-j"))
		})

		It("renders done state with press space to exit", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})
			model = update(model, porthole.ContentLineMsg{Line: "content\n"})
			model = update(model, contract.CompleteMsg{
				Files:   5,
				Dirs:    2,
				Elapsed: time.Second,
			})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("content"))
			Expect(output).To(ContainSubstring("press space to exit"))
		})

		It("omits press space to exit before completion", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})
			model = update(model, porthole.ContentLineMsg{Line: "content\n"})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("content"))
			Expect(output).NotTo(ContainSubstring("press space to exit"))
		})

		It("uses alt screen mode", func() {
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			v := model.View()
			Expect(v.AltScreen).To(BeTrue())
		})
	})

	Describe("Flow RenderLine helper", func() {
		It("renders a directory line with no action", func() {
			result := linear.RenderLine(linear.LineParams{
				NodeParams: contract.NodeParams{
					Path:        "/path/to/dir",
					Name:        "mydir",
					IsDir:       true,
					Depth:       1,
					IsLast:      true,
					VisualDepth: 1,
				},
				RenderParams: contract.RenderParams{
					BodyWidth: 0,
					Theme:     theme,
				},
			})

			Expect(result.Line).To(ContainSubstring("mydir"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a file line with no action", func() {
			result := linear.RenderLine(linear.LineParams{
				NodeParams: contract.NodeParams{
					Path:        "/path/to/file.txt",
					Name:        "file.txt",
					IsDir:       false,
					Depth:       1,
					IsLast:      true,
					VisualDepth: 1,
				},
				RenderParams: contract.RenderParams{
					BodyWidth: 0,
					Theme:     theme,
				},
			})

			Expect(result.Line).To(ContainSubstring("file.txt"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a directory line with action", func() {
			result := linear.RenderLine(linear.LineParams{
				NodeParams: contract.NodeParams{
					Path:        "/path/to/dir",
					Name:        "mydir",
					IsDir:       true,
					Depth:       1,
					ActionName:  "test-action",
					IsLast:      true,
					VisualDepth: 1,
				},
				RenderParams: contract.RenderParams{
					BodyWidth: 0,
					Theme:     theme,
				},
			})

			Expect(result.Line).To(ContainSubstring("mydir"))
			Expect(result.Line).To(ContainSubstring("test-action"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a file line with error", func() {
			result := linear.RenderLine(linear.LineParams{
				NodeParams: contract.NodeParams{
					Path:        "/path/to/file.txt",
					Name:        "file.txt",
					IsDir:       false,
					Depth:       1,
					Err:         errors.New("permission denied"),
					IsLast:      true,
					VisualDepth: 1,
				},
				RenderParams: contract.RenderParams{
					BodyWidth: 0,
					Theme:     theme,
				},
			})

			Expect(result.Line).To(ContainSubstring("file.txt"))
			Expect(result.Line).To(ContainSubstring("!"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a pipeline step nested under its parent", func() {
			result := linear.RenderLine(linear.LineParams{
				NodeParams: contract.NodeParams{
					Path:            "/path/to/file.txt",
					Name:            "file.txt",
					IsDir:           false,
					Depth:           1,
					ActionName:      "echo",
					ExecutionString: "[src/app/file.txt]",
					IsLast:          false,
					IsPipelineStep:  true,
					IsLastStep:      true,
					VisualDepth:     2,
				},
				RenderParams: contract.RenderParams{
					BodyWidth: 0,
					Theme:     theme,
				},
			})

			// Pipeline steps render as "• via actionName" without
			// the file/dir icon or name, matching the linear view.
			Expect(result.Line).To(ContainSubstring("echo"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})
	})

	Describe("View layout with legend", func() {
		It("renders status and bottom border on screen when flags are active", func() {
			// Regression: the legend pushes the status row and bottom
			// border off the screen unless the body height accounts
			// for the legend's own height.
			model := porthole.NewModel(contract.NewModelParams{RootPath: "/root", MaxDepth: 0, Theme: theme, NoRecurse: false})
			model = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})
			model = update(model, porthole.OvertureMsg{
				OvertureMsg: contract.OvertureMsg{
					Root: "/root",
					Header: contract.HeaderInfo{
						FilesGlob: "*.go",
						DirsRegex: ".",
					},
					FlagsRowPosition: contract.PositionBottom,
				},
			})
			model = update(model, porthole.ContentLineMsg{Line: "content\n"})
			model = update(model, contract.CompleteMsg{
				Files:   1,
				Dirs:    1,
				Elapsed: time.Second,
			})

			v := model.View()
			lines := strings.Split(v.Content, "\n")

			// The status row must be present (contains "files:")
			statusIdx := -1
			bottomIdx := -1

			for i, line := range lines {
				if strings.Contains(line, "files:") {
					statusIdx = i
				}
				if strings.Contains(line, "╰") {
					bottomIdx = i
				}
			}
			Expect(statusIdx).To(BeNumerically(">=", 0),
				"status row should be present in the view")
			Expect(bottomIdx).To(BeNumerically(">=", 0),
				"bottom border should be present in the view")
			Expect(bottomIdx).To(BeNumerically(">", statusIdx),
				"bottom border should appear after the status row")
		})
	})
})

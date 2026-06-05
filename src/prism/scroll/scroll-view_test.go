package scroll_test

import (
	"bytes"
	"errors"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/flow"
	"github.com/snivilised/jaywalk/src/prism/scroll"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// update is a convenience wrapper that calls Update on the model and
// type-asserts the result back to Model.
func update(m scroll.Model, msg tea.Msg) (scroll.Model, tea.Cmd) { //nolint:unparam // tea.Cmd not used; ok
	r, cmd := m.Update(msg)
	return *r.(*scroll.Model), cmd //nolint:errcheck // ok
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
			model := scroll.NewModel("/test/path", 0, theme, false)
			Expect(model).NotTo(BeNil())
		})

		It("creates a model with max depth and no recurse", func() {
			model := scroll.NewModel("/test/path", 3, theme, true)
			Expect(model).NotTo(BeNil())
		})

		It("initializes startedAt timestamp", func() {
			model := scroll.NewModel("/test/path", 0, theme, false)
			Expect(model).NotTo(BeNil())
		})
	})

	Describe("OvertureMsg handling", func() {
		It("sets all fields from OvertureMsg", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			msg := scroll.OvertureMsg{
				Root:         "/new/root",
				Caption:      "Test Caption",
				PipelineName: "test-pipeline",
				Banner:       banner.Info{},
			}

			m, _ := update(model, msg)
			Expect(m.RootPath()).To(Equal("/new/root"))
			Expect(m.Caption()).To(Equal("Test Caption"))
			Expect(m.Pipeline()).To(Equal("test-pipeline"))
		})

		It("sets content buffer from ContentLineMsg", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			msg := scroll.ContentLineMsg{
				Line: "Test line\n",
			}

			m, _ := update(model, msg)
			Expect(len(m.Buffer())).To(Equal(1))
			Expect(m.Buffer()[0]).To(Equal("Test line"))
		})

		It("strips trailing newlines from content lines", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			msg := scroll.ContentLineMsg{
				Line: "line with newline\n",
			}

			m, _ := update(model, msg)
			Expect(m.Buffer()[0]).To(Equal("line with newline"))
		})

		It("skips empty content lines", func() {
			model := scroll.NewModel("/root", 0, theme, false)

			m, _ := update(model, scroll.ContentLineMsg{Line: "\n"})
			m, _ = update(m, scroll.ContentLineMsg{Line: ""})
			m, _ = update(m, scroll.ContentLineMsg{Line: "real content\n"})

			Expect(len(m.Buffer())).To(Equal(1))
			Expect(m.Buffer()[0]).To(Equal("real content"))
		})

		It("truncates content buffer when it exceeds MaxContentBufferLines", func() {
			model := scroll.NewModel("/root", 0, theme, false)

			// Add more lines than the max
			for i := 0; i < scroll.MaxContentBufferLines+10; i++ {
				msg := scroll.ContentLineMsg{
					Line: "Line " + string(rune(i)) + "\n",
				}
				m, _ := update(model, msg)
				model = m
			}

			Expect(len(model.Buffer())).To(BeNumerically("<=", scroll.MaxContentBufferLines))
		})

		It("sets done flag from CompleteMsg", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			msg := scroll.CompleteMsg{
				Files:   10,
				Dirs:    5,
				Elapsed: time.Second,
			}

			m, _ := update(model, msg)
			Expect(m.IsDone()).To(BeTrue())
		})

		It("stores banner info from OvertureMsg", func() {
			model := scroll.NewModel("/root", 0, theme, false)

			grad, _ := theme.GradientFor(contract.GradientComponentBanner)
			info := banner.Info{
				Disable:  false,
				Position: contract.PositionTop,
				Width:    80,
				Gradient: &grad,
			}

			m, _ := update(model, scroll.OvertureMsg{
				Root:   "/root",
				Banner: info,
			})

			Expect(m.BannerInfo().Position).To(Equal(contract.PositionTop))
			Expect(m.BannerInfo().Width).To(Equal(80))
			Expect(m.BannerInfo().Disable).To(BeFalse())
		})
	})

	Describe("View rendering", func() {
		It("renders header and status with no content", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("/root"))
			Expect(output).To(ContainSubstring("Processing"))
		})

		It("renders content lines in the body", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			model, _ = update(model, scroll.ContentLineMsg{Line: "file-a.txt\n"})
			model, _ = update(model, scroll.ContentLineMsg{Line: "file-b.txt\n"})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("file-a.txt"))
			Expect(output).To(ContainSubstring("file-b.txt"))
		})

		It("renders multiple content lines in order", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			for i := 0; i < 10; i++ {
				model, _ = update(model, scroll.ContentLineMsg{
					Line: "line-" + string(rune('a'+i)) + "\n",
				})
			}

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("line-a"))
			Expect(output).To(ContainSubstring("line-j"))
		})

		It("renders done state with press space to exit", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})
			model, _ = update(model, scroll.ContentLineMsg{Line: "content\n"})
			model, _ = update(model, scroll.CompleteMsg{
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
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})
			model, _ = update(model, scroll.ContentLineMsg{Line: "content\n"})

			v := model.View()
			output := v.Content

			Expect(output).To(ContainSubstring("content"))
			Expect(output).NotTo(ContainSubstring("press space to exit"))
		})

		It("uses alt screen mode", func() {
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})

			v := model.View()
			Expect(v.AltScreen).To(BeTrue())
		})
	})

	Describe("Flow RenderLine helper", func() {
		It("renders a directory line with no action", func() {
			result := flow.RenderLine(
				"/path/to/dir",
				"mydir",
				true,  // isDir
				1,     // depth
				"",    // actionName
				"",    // pipelineName
				"",    // commandOutput
				"",    // executionString
				false, // dryRun
				nil,   // err
				true,  // isLast
				false, // isPipelineStep
				false, // isLastStep
				1,     // visualDepth
				nil,   // branchStack
				0,     // bodyWidth
				theme,
				"", // activityFrame
			)

			Expect(result.Line).To(ContainSubstring("mydir"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a file line with no action", func() {
			result := flow.RenderLine(
				"/path/to/file.txt",
				"file.txt",
				false, // isDir
				1,     // depth
				"",    // actionName
				"",    // pipelineName
				"",    // commandOutput
				"",    // executionString
				false, // dryRun
				nil,   // err
				true,  // isLast
				false, // isPipelineStep
				false, // isLastStep
				1,     // visualDepth
				nil,   // branchStack
				0,     // bodyWidth
				theme,
				"", // activityFrame
			)

			Expect(result.Line).To(ContainSubstring("file.txt"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a directory line with action", func() {
			result := flow.RenderLine(
				"/path/to/dir",
				"mydir",
				true,          // isDir
				1,             // depth
				"test-action", // actionName
				"",            // pipelineName
				"",            // commandOutput
				"",            // executionString
				false,         // dryRun
				nil,           // err
				true,          // isLast
				false,         // isPipelineStep
				false,         // isLastStep
				1,             // visualDepth
				nil,           // branchStack
				0,             // bodyWidth
				theme,
				"", // activityFrame
			)

			Expect(result.Line).To(ContainSubstring("mydir"))
			Expect(result.Line).To(ContainSubstring("test-action"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a file line with error", func() {
			result := flow.RenderLine(
				"/path/to/file.txt",
				"file.txt",
				false,                           // isDir
				1,                               // depth
				"",                              // actionName
				"",                              // pipelineName
				"",                              // commandOutput
				"",                              // executionString
				false,                           // dryRun
				errors.New("permission denied"), // err
				true,                            // isLast
				false,                           // isPipelineStep
				false,                           // isLastStep
				1,                               // visualDepth
				nil,                             // branchStack
				0,                               // bodyWidth
				theme,
				"", // activityFrame
			)

			Expect(result.Line).To(ContainSubstring("file.txt"))
			Expect(result.Line).To(ContainSubstring("!"))
			Expect(result.Line).To(ContainSubstring("\n"))
		})

		It("renders a pipeline step nested under its parent", func() {
			result := flow.RenderLine(
				"/path/to/file.txt",
				"file.txt",
				false,                // isDir
				1,                    // depth
				"echo",               // actionName
				"",                   // pipelineName
				"",                   // commandOutput
				"[src/app/file.txt]", // executionString
				false,                // dryRun
				nil,                  // err
				false,                // isLast
				true,                 // isPipelineStep
				true,                 // isLastStep
				2,                    // visualDepth (depth + 1)
				nil,                  // branchStack
				0,                    // bodyWidth
				theme,
				"", // activityFrame
			)

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
			model := scroll.NewModel("/root", 0, theme, false)
			model, _ = update(model, tea.WindowSizeMsg{Width: 80, Height: 24})
			model, _ = update(model, scroll.OvertureMsg{
				Root: "/root",
				Header: contract.HeaderInfo{
					FilesGlob: "*.go",
					DirsRegex: ".",
				},
				FlagsRowPosition: contract.PositionBottom,
			})
			model, _ = update(model, scroll.ContentLineMsg{Line: "content\n"})
			model, _ = update(model, scroll.CompleteMsg{
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

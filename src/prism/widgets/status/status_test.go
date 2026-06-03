package status_test

import (
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/status"
)

func TestStatus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Status Suite")
}

var _ = Describe("Status", func() {
	var (
		styles status.Styles
	)

	BeforeEach(func() {
		styles = status.Styles{
			TreeIcons: contract.TreeIcons{
				contract.TreeIconFile:      "🔖",
				contract.TreeIconDirectory: "📁",
				contract.TreeIconError:     "🚫",
				contract.TreeIconSkipped:   "⛔️",
				contract.TreeIconElapsed:   "⏰",
			},
			SummaryLabelStyle: lipgloss.NewStyle(),
			SummaryValueStyle: lipgloss.NewStyle(),
			ErrorStyle:        lipgloss.NewStyle(),
			ProgressStyle:     lipgloss.NewStyle(),
			BorderStyle:       lipgloss.NewStyle(),
			MutedStyle:        lipgloss.NewStyle(),
		}
	})

	Describe("Render", func() {
		It("renders all fields when all are enabled", func() {
			fields := status.FieldSelectors{
				ShowFiles:    true,
				ShowDirs:     true,
				ShowErrors:   true,
				ShowSkipped:  true,
				ShowProgress: false,
				ShowComplete: false,
				ShowElapsed:  true,
			}

			output := status.Render(status.Config{
				Files:   42,
				Dirs:    7,
				Errors:  0,
				Skipped: 3,
				Elapsed: 5 * time.Second,
			}, styles, fields, 80)

			Expect(output).To(ContainSubstring("🔖 files:"))
			Expect(output).To(ContainSubstring("📁 dirs:"))
			Expect(output).To(ContainSubstring("🚫 errors:"))
			Expect(output).To(ContainSubstring("⛔️ skipped:"))
			Expect(output).To(ContainSubstring("⏰ elapsed:"))
		})

		It("renders only enabled fields", func() {
			fields := status.FieldSelectors{
				ShowFiles:    true,
				ShowDirs:     false,
				ShowErrors:   false,
				ShowSkipped:  false,
				ShowProgress: false,
				ShowComplete: false,
				ShowElapsed:  true,
			}

			output := status.Render(status.Config{
				Files:   42,
				Dirs:    7,
				Errors:  0,
				Skipped: 3,
				Elapsed: 5 * time.Second,
			}, styles, fields, 80)

			Expect(output).To(ContainSubstring("🔖 files:"))
			Expect(output).To(ContainSubstring("⏰ elapsed:"))
			Expect(output).ToNot(ContainSubstring("📁 dirs:"))
			Expect(output).ToNot(ContainSubstring("🚫 errors:"))
			Expect(output).ToNot(ContainSubstring("⛔️ skipped:"))
		})

		It("formats duration correctly", func() {
			fields := status.FieldSelectors{
				ShowFiles:    false,
				ShowDirs:     false,
				ShowErrors:   false,
				ShowSkipped:  false,
				ShowProgress: false,
				ShowComplete: false,
				ShowElapsed:  true,
			}

			output := status.Render(status.Config{
				Elapsed: 90 * time.Second,
			}, styles, fields, 80)

			Expect(output).To(ContainSubstring("1m30s"))
		})

		It("formats milliseconds for sub-second durations", func() {
			fields := status.FieldSelectors{
				ShowElapsed: true,
			}

			output := status.Render(status.Config{
				Elapsed: 500 * time.Millisecond,
			}, styles, fields, 80)

			Expect(output).To(ContainSubstring("500ms"))
		})

		It("includes border characters", func() {
			fields := status.FieldSelectors{
				ShowFiles:   true,
				ShowElapsed: true,
			}

			output := status.Render(status.Config{
				Files:   42,
				Elapsed: 5 * time.Second,
			}, styles, fields, 80)

			Expect(output).To(ContainSubstring("│"))
		})
	})
})

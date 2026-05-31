package widget

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/snivilised/jaywalk/src/agenor/core"

	"charm.land/lipgloss/v2"
)

var _ = Describe("DateTime", func() {
	It("returns empty string for empty subscription label", func() {
		styles := DateTimeStyles{InfoStyle: lipgloss.NewStyle()}
		result := DateTime("", core.Now(), "", styles)
		Expect(result).To(Equal(""))
	})

	It("returns empty string for zero time", func() {
		styles := DateTimeStyles{InfoStyle: lipgloss.NewStyle()}
		result := DateTime("label", time.Time{}, "", styles)
		Expect(result).To(Equal(""))
	})

	It("formats with default date format if not provided", func() {
		styles := DateTimeStyles{InfoStyle: lipgloss.NewStyle()}
		now := time.Date(2026, time.January, 2, 15, 4, 5, 0, time.FixedZone("MST", -7*60*60))
		expect := fmt.Sprintf("  %s  -  %s", "test-label", now.Format("Mon, 02 Jan 2006 15:04:05 MST"))
		result := DateTime("test-label", now, "", styles)
		Expect(result).To(Equal(styles.InfoStyle.Render(expect)))
	})

	It("formats with custom date format", func() {
		styles := DateTimeStyles{InfoStyle: lipgloss.NewStyle()}
		now := time.Date(2026, time.January, 2, 15, 4, 5, 0, time.UTC)
		dateFormat := "2006-01-02 15:04"
		expect := fmt.Sprintf("  %s  -  %s", "test-label", now.Format(dateFormat))
		result := DateTime("test-label", now, dateFormat, styles)
		Expect(result).To(Equal(styles.InfoStyle.Render(expect)))
	})

	It("applies InfoStyle", func() {
		infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
		styles := DateTimeStyles{InfoStyle: infoStyle}
		now := core.Now()
		expect := fmt.Sprintf("  %s  -  %s", "label", now.Format("Mon, 02 Jan 2006 15:04:05 MST"))
		result := DateTime("label", now, "", styles)
		Expect(result).To(Equal(infoStyle.Render(expect)))
	})
})

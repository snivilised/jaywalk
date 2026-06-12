package shared

import (
	"strings"

	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
	"github.com/snivilised/jaywalk/src/prism/widgets/legend"
)

// WriteBanner constructs a transient banner.Model from info and width,
// checks whether it is disabled, and writes its rendered output to b.
// The caller is responsible for emitting any separator newline that
// must precede the banner (e.g. when rendering at the bottom position).
// When the banner is disabled or produces no output, b is unchanged.
func WriteBanner(b *strings.Builder, info banner.Info, width int) {
	bm := banner.NewModel(
		banner.WithInfo(info),
		banner.WithWidth(width),
	)
	if bm.Disabled() {
		return
	}
	if out := bm.View(); out != "" {
		b.WriteString(out)
	}
}

// WriteLegend constructs a transient legend.Model from the supplied
// parameters and writes its rendered output to b. When no flags are
// active the legend produces no output and b is unchanged.
func WriteLegend(b *strings.Builder, params LegendParams) {
	lm := newLegendModel(params)
	if out := lm.View(); out != "" {
		b.WriteString(out)
	}
}

// LegendHeight returns the number of terminal rows the legend will
// occupy (entry lines only). Returns 0 when no flags are active so
// the caller can skip the section entirely.
func LegendHeight(params LegendParams) int {
	return newLegendModel(params).Height()
}

// LegendSectionHeight returns the total number of rows the legend
// section occupies: entry lines plus the two separator borders that
// frame it (one above, one below). Returns 0 when no flags are active
// so callers can skip the section and claim the full available height.
func LegendSectionHeight(params LegendParams) int {
	h := LegendHeight(params)
	if h == 0 {
		return 0
	}
	return h + 2
}

// LegendParams groups the inputs required to construct a transient
// legend.Model. Both the highway and porthole views populate this on
// every render; the legend widget itself is stateless.
type LegendParams struct {
	Info   legend.Info
	Width  int
	Styles legend.Styles
}

// newLegendModel is the single construction site for the transient
// legend.Model. All exported functions in this file delegate here so
// the construction options stay consistent.
func newLegendModel(p LegendParams) legend.Model {
	return legend.NewModel(
		legend.WithInfo(p.Info),
		legend.WithWidth(p.Width),
		legend.WithStyles(p.Styles),
	)
}

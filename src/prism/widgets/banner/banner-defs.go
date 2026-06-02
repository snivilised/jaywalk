package banner

import "time"

// ---------------------------------------------------------------------------
// Banner position
// ---------------------------------------------------------------------------
//
// The banner is rendered OUTSIDE the bordered region of the host view.
// These constants therefore do not interact with the highway view's
// FlagsRowPosition - the flags row lives inside the border, the banner
// lives outside it.

const (
	// PositionTop places the banner above the top border.
	PositionTop = "top"

	// PositionBottom places the banner below the summary (and thus
	// below the bottom border).
	PositionBottom = "bottom"
)

// ---------------------------------------------------------------------------
// Justification
// ---------------------------------------------------------------------------

const (
	// JustifyRight aligns the right edge of every banner line with
	// the right edge of the terminal. This is the default.
	JustifyRight = "right"

	// JustifyLeft preserves the banner's intrinsic leading spaces
	// without further padding. The art is therefore rendered flush
	// against the left of the host view.
	JustifyLeft = "left"

	// JustifyCenter centres every banner line within the terminal.
	JustifyCenter = "center"
)

// ---------------------------------------------------------------------------
// Character classes
// ---------------------------------------------------------------------------

// faceRune is the FULL BLOCK glyph (U+2588). It is the only character
// classified as a "face" rune for the purposes of the unity aspect.
const faceRune = '█'

// ---------------------------------------------------------------------------
// Animation timing
// ---------------------------------------------------------------------------

// DefaultBannerTick is the tick interval (milliseconds) used by the
// banner animation when the user has not configured one. It is
// deliberately slower than the highway's global tick (~50ms) so the
// gradient sweep reads as a warm glow rather than a strobe.
const DefaultBannerTick = 500 * time.Millisecond

// ---------------------------------------------------------------------------
// Default art
// ---------------------------------------------------------------------------

// DefaultArt is the JAYWALK ASCII banner used when the caller does not
// provide their own Art. It contains 6 lines rendered with FULL BLOCK
// (█) face runes and box-drawing shadow runes (╔╗╚╝═║). The art is
// designed for a 24-rune-wide right-justified presentation against a
// typical 80-column terminal. Each line is prefixed with a single space
// so the inner face characters start in column 1.
//
// Callers can override the art by setting Config.Art.
const DefaultArt = `
     ██╗ █████╗ ██╗   ██╗██╗    ██╗ █████╗ ██╗     ██╗  ██╗
     ██║██╔══██╗╚██╗ ██╔╝██║    ██║██╔══██╗██║     ██║ ██╔╝
     ██║███████║ ╚████╔╝ ██║ █╗ ██║███████║██║     █████╔╝ 
██   ██║██╔══██║  ╚██╔╝  ██║███╗██║██╔══██║██║     ██╔═██╗ 
╚█████╔╝██║  ██║   ██║   ╚███╔███╔╝██║  ██║███████╗██║  ██╗
 ╚════╝ ╚═╝  ╚═╝   ╚═╝    ╚══╝╚══╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝
`

package highway

import (
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

type OvertureMsg struct {
	contract.OvertureMsg

	ActionName string

	// Banner carries the optional ANSI shadow banner info. The
	// highway model uses banner.Info directly (no highway-specific
	// wrapper type). When the Disable flag is true or the gradient
	// is nil, the view skips it.
	Banner banner.Info
}

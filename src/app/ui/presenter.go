package ui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/prism/contract"
	"github.com/snivilised/jaywalk/src/prism/widgets/banner"
)

// presenter is the base presenter struct that contains common state and
// functionality shared by highwayPresenter and portholePresenter. It manages
// the bubbletea program lifecycle, theme, and common event handling.
type presenter struct {
	program    *tea.Program
	done       chan struct{}
	noW        uint
	maxDepth   uint
	totalFiles uint
	totalDirs  uint
	theme      contract.Theme
	noRecurse  bool

	// header is the supplementary flag info carried on the BeginEvent.
	// Stored on the presenter so tests and lifecycle hooks can introspect
	// it, and read here when building the OvertureMsg. See
	// contract.HeaderInfo for field semantics.
	header contract.HeaderInfo

	// bannerInfo is built once in OnBegin and sent to the model via
	// OvertureMsg. It carries the random aspects and gradient
	// endpoints resolved from the theme.
	bannerInfo banner.Info
}

func (p *presenter) OnTraversalOptions(o *pref.Options) {
	// Read concurrency settings from options - this is structural configuration, not cascade state.
	p.noW = o.Concurrency.NoW
}

func (p *presenter) SetMaxDepth(maxDepth uint) {
	p.maxDepth = maxDepth
}

func (p *presenter) NeedsPeerInfo() bool {
	return true
}

func (p *presenter) OnPeerInfoBegin(files, dirs uint, _ map[string]*core.PeerInfo) {
	p.totalFiles = files
	p.totalDirs = dirs
}

func (p *presenter) OnPeerInfoEnd() {}

func subscriptionLabelFor(s enums.Subscription) string {
	switch s {
	case enums.SubscribeUndefined:
		return "undefined"
	case enums.SubscribeFiles:
		return "files only"
	case enums.SubscribeDirectories:
		return "folders only"
	case enums.SubscribeDirectoriesWithFiles:
		return "directories w/ files"
	case enums.SubscribeUniversal:
		return "files and folders"
	default:
		return "files and folders"
	}
}

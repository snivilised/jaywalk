package legend

import "github.com/snivilised/jaywalk/src/prism/contract"

// Info carries the per-render and per-session flags-row configuration.
// The presenter builds one of these once in OnBegin; the highway model
// stores the underlying contract.HeaderInfo and the position string
// and reads from it on every render.
//
// Position: contract.PositionTop | contract.PositionBottom | ""
// (empty / unrecognised values mean "do not render").
// Header: the per-session flag data (cascade, filter, sampler).
type Info struct {
	Position string
	Header   contract.HeaderInfo
}

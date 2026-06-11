package ansi

import (
	"fmt"

	"github.com/snivilised/jaywalk/src/prism/contract"
)

func EscapedTrueColor(r, g, b uint8) string {
	return fmt.Sprintf("%s%s;%s;%d;%d;%dm",
		contract.Static.ANSIEscape.Escape,
		contract.Static.ANSIEscape.ForeGroundColor,
		contract.Static.ANSIEscape.TrueColor,
		r, g, b,
	)
}

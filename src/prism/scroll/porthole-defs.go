package scroll

import (
	"time"
)

// MaxContentBufferLines is the hard cap on content lines stored in the
// porthole's buffer. The view truncates from the front when this limit
// is exceeded, dropping ContentBufferTruncateStep lines at a time to
// keep navigation responsive.
const MaxContentBufferLines = 2000

// ContentBufferTruncateStep is how many lines are dropped from the
// front of the buffer in one truncation cycle. The step size is
// deliberately larger than a single line so that truncation does not
// happen on every render when the buffer grows; it only triggers at
// multiples of this step.
const ContentBufferTruncateStep = 100

// TickRate is the interval between UI ticks that advance the banner
// gradient and update elapsed time. It mirrors highway's tick rate so
// both views have consistent timing for their animations.
var TickRate = 50 * time.Millisecond

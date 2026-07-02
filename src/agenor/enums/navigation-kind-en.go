package enums

import "fmt"

//go:generate stringer -type=NavigationKind -linecomment -trimprefix=NavigationKind -output navigation-kind-en-auto.go

// NavigationKind identifies whether a traversal is fresh or a
// continuation from a checkpoint.
type NavigationKind uint

const (
	// NavigationKindPrime is a fresh traversal from the root.
	NavigationKindPrime NavigationKind = iota // prime

	// NavigationKindResume is a continuation from a saved checkpoint.
	NavigationKindResume // resume
)

// UnmarshalText implements encoding.TextUnmarshaler so that NavigationKind
// can be decoded from string values.
func (k *NavigationKind) UnmarshalText(data []byte) error {
	switch string(data) {
	case "prime":
		*k = NavigationKindPrime
	case "resume":
		*k = NavigationKindResume
	default:
		return fmt.Errorf("unknown navigation kind %q: valid values are prime, resume", string(data))
	}
	return nil
}

package contract

// NavigationKind identifies whether a traversal is fresh or a
// continuation from a checkpoint.
type NavigationKind string

const (
	// PrimeNavigation is a fresh traversal from the root.
	PrimeNavigation NavigationKind = "prime"

	// ResumeNavigation is a continuation from a saved checkpoint.
	ResumeNavigation NavigationKind = "resume"
)

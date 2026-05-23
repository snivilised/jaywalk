package pref

type (
	// RescueData is the data to be saved in case of a rescue.
	RescueData interface {
		// Data returns the data to be saved.
		Data() any
	}

	// Recovery is the interface for saving rescue data.
	Recovery interface {
		// Save saves the rescue data and returns the path write to.
		Save(data RescueData) (string, error)
	}

	// TraversalConfigurer used to enable the ui to modify the traversal
	// for its own requirement.
	TraversalConfigurer interface {
		// OnTraversalOptions invoked when options are available to be modified
		OnTraversalOptions(o *Options)
	}
)

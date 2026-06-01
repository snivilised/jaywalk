package command

import (
	"fmt"
	"strings"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/controller"
	"github.com/snivilised/jaywalk/src/locale"
)

// Poly filter flag names (must match the first token of the i18n
// usage text used to bind them in store.PolyFilterParameterSet).
const (
	polyFlagFiles      = "files"
	polyFlagFilesRegex = "files-regex"
	polyFlagDirsGlob   = "dirs-glob"
	polyFlagDirsRegex  = "dirs-regex"
)

// createTraversalSettingsIntent builds a TraversalSettingsIntent from the
// nav families. Poly filter fields are populated only when the user
// actually set the corresponding flag on the command line (detected via
// cobra's Flags().Changed), so an empty string in the resulting
// FilterIntent reliably means "user did not specify this filter" -
// independent of any default values the flag may carry.
func createTraversalSettingsIntent(families NavFamilies) controller.TraversalSettingsIntent {
	intent := controller.TraversalSettingsIntent{
		NoRecurse:     families.Cascade.Native.NoRecurse,
		Depth:         core.TraversalDepth(families.Cascade.Native.Depth),
		IsSampling:    families.Sampling.Native.IsSampling,
		NoFiles:       families.Sampling.Native.NoFiles,
		NoDirectories: families.Sampling.Native.NoDirectories,
	}

	flags := families.PolyFam.Command.Flags()
	native := families.PolyFam.Native

	if flags.Changed(polyFlagFiles) {
		intent.Filter.FilesExGlob = native.FilesExGlob
	}
	if flags.Changed(polyFlagFilesRegex) {
		intent.Filter.FilesRegEx = native.FilesRegEx
	}
	if flags.Changed(polyFlagDirsGlob) {
		intent.Filter.DirectoriesGlob = native.DirectoriesGlob
	}
	if flags.Changed(polyFlagDirsRegex) {
		intent.Filter.DirectoriesRegEx = native.DirectoriesRegEx
	}

	return intent
}

// resolveResumeStrategy maps the --resume flag string to the agenor constant.
func resolveResumeStrategy(resume string) (agenor.ResumeStrategy, error) {
	switch resume {
	case ResumeStrategySpawn:
		return agenor.ResumeStrategySpawn, nil
	case ResumeStrategyFastward:
		return agenor.ResumeStrategyFastward, nil
	default:
		return 0, locale.NewInvalidResumeValueError(
			resume,
			fmt.Sprintf("'%s'", strings.Join([]string{
				ResumeStrategySpawn, ResumeStrategyFastward,
			}, ", ")),
		)
	}
}

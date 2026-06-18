package controller

import (
	"fmt"
	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/filing"
	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/locale"
)

type FilterIntent struct {
	FilesExGlob      string
	FilesRegEx       string
	DirectoriesGlob  string
	DirectoriesRegEx string
}

type TraversalSettingsIntent struct {
	NoRecurse     bool
	Depth         core.TraversalDepth
	IsSampling    bool
	NoFiles       uint
	NoDirectories uint
	Filter        FilterIntent
	IncludeHidden bool
}

func BuildTraversalSettings(intent TraversalSettingsIntent, ui report.Presenter) []pref.Option {
	var opts []pref.Option

	if intent.NoRecurse {
		opts = append(opts, agenor.WithNoRecurse())
	}

	if intent.Depth > 0 {
		opts = append(opts, agenor.WithDepth(intent.Depth))
	}

	if intent.IsSampling {
		opts = append(opts, agenor.WithSamplingOptions(&pref.SamplingOptions{
			NoOf: pref.EntryQuantities{
				Files:       intent.NoFiles,
				Directories: intent.NoDirectories,
			},
		}))
	}

	if filterOption, ok := TranslateFilterIntent(intent.Filter); ok {
		opts = append(opts, filterOption)
	}

	if intent.IncludeHidden {
		opts = append(opts, agenor.WithHookReadDirectory(filing.ReadEntriesAll))
	}

	opts = append(opts, pref.WithTraversalConfigurer(ui))

	return opts
}

func TranslateFilterIntent(intent FilterIntent) (pref.Option, bool) {
	hasFilesFlag := intent.FilesExGlob != "" || intent.FilesRegEx != ""
	hasDirectoriesFlag := intent.DirectoriesGlob != "" || intent.DirectoriesRegEx != ""

	if !hasFilesFlag && !hasDirectoriesFlag {
		return nil, false
	}

	// Build file filter definition with proper type detection
	var fileDef core.FilterDef
	fileHasFilter := intent.FilesExGlob != "" || intent.FilesRegEx != ""

	if fileHasFilter {
		switch {
		case intent.FilesExGlob != "":
			fileDef.Type = enums.FilterTypeGlobEx
			fileDef.Pattern = intent.FilesExGlob
							fileDef.Description = fmt.Sprintf("files-glob:%s", intent.FilesExGlob)
				fileDef.IfNotApplicable = enums.TriStateBoolFalse
		case intent.FilesRegEx != "":
			fileDef.Type = enums.FilterTypeRegex
			fileDef.Pattern = intent.FilesRegEx
							fileDef.Description = fmt.Sprintf("files-regex:%s", intent.FilesRegEx)
				fileDef.IfNotApplicable = enums.TriStateBoolFalse
		}
	} else {
		fileDef = core.BenignNodeFilterDef // allow all files when no file filter specified
	}

	// Build directory filter definition with proper type detection
	var dirDef core.FilterDef
	dirHasFilter := intent.DirectoriesGlob != "" || intent.DirectoriesRegEx != ""

	if dirHasFilter {
		switch {
		case intent.DirectoriesGlob != "":
			dirDef.Type = enums.FilterTypeGlob
			dirDef.Pattern = intent.DirectoriesGlob
							dirDef.Description = fmt.Sprintf("dirs-glob:%s", intent.DirectoriesGlob)
				dirDef.IfNotApplicable = enums.TriStateBoolFalse
		case intent.DirectoriesRegEx != "":
			dirDef.Type = enums.FilterTypeRegex
			dirDef.Pattern = intent.DirectoriesRegEx
							dirDef.Description = fmt.Sprintf("dirs-regex:%s", intent.DirectoriesRegEx)
				dirDef.IfNotApplicable = enums.TriStateBoolFalse
		}
	} else {
		dirDef = core.BenignNodeFilterDef // allow all directories when no dir filter specified
	}

	return pref.WithFilter(&pref.FilterOptions{
		Node: &core.FilterDef{
			Type: enums.FilterTypePoly,
			Poly: &core.PolyFilterDef{
				File:      fileDef,
				Directory: dirDef,
			},
		},
	}), true
}

func ResolveSubscription(flag string) (enums.Subscription, error) {
	switch flag {
	case "files", "":
		return enums.SubscribeFiles, nil
	case "dirs":
		return enums.SubscribeDirectories, nil
	case "all":
		return enums.SubscribeUniversal, nil
	default:
		return 0, locale.ErrInvalidSubscription
	}
}

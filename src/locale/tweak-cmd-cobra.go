package locale

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// TweakCmdLongDescTemplData interactive theme editor.
type TweakCmdLongDescTemplData struct {
	agenorTemplData
}

// Message returns the i18n message for TweakCmdLongDescTemplData.
func (td TweakCmdLongDescTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "tweak-command-long-description",
		Description: "interactive theme editor",
		Other: `Interactive theme editor and gradient workshop. Provides a
terminal UI for creating and editing jay colour themes with
live preview rendering.`,
	}
}

// TweakCmdShortDescTemplData interactive theme editor.
type TweakCmdShortDescTemplData struct {
	agenorTemplData
}

// Message returns the i18n message for TweakCmdShortDescTemplData.
func (td TweakCmdShortDescTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "tweak-command-short-description",
		Description: "interactive theme editor",
		Other:       "Interactive theme editor with live preview",
	}
}

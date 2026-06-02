package contract

type (
	EmojiData struct {
		// Padlock is the cascade widget value used across the
		// flags-row specs. Lifted into a constant to satisfy the
		// goconst linter check for repeated string literals.
		Padlock string
		Snail   string
	}
	BordersData struct {
		TopLeftCorner     string
		TopRight          string
		BottomLeft        string
		BottomLeftCorner  string
		BottomRightCorner string
	}
	StaticData struct {
		Emoji   EmojiData
		Borders BordersData
	}
)

var (
	Static = StaticData{
		Emoji: EmojiData{
			Padlock: "🔒",
			Snail:   "🐌",
		},
		Borders: BordersData{
			TopLeftCorner:     "╭",
			TopRight:          ".★..─╮",
			BottomLeft:        "╰─..★.",
			BottomLeftCorner:  "╰",
			BottomRightCorner: "╯",
		},
	}
)

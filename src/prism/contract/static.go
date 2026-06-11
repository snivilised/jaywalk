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
	// ANSIEscapeData
	// https://en.wikipedia.org/wiki/ANSI_escape_code
	// Expected sequence for 'A': \x1b[38;2;255;128;64mA\x1b[0m
	// \x1b -> escape
	// [ -> Control Sequence Introducer (CSI) - starts an ANSI escape sequence
	// 38 -> foreground colour
	// 48 -> background colour
	// TrueColor (24-bit) -> ;2;{r};{g};{b}
	// 256-color -> ;5;{n}
	// m  -> ends the SGR command
	// A -> is the character being printed with the specified colour
	ANSIEscapeData struct {
		Escape          string
		Reset           string
		TrueColor       string
		TrueColorFormat string
		ForeGroundColor string
		BackGroundColor string
	}
	StaticData struct {
		Emoji      EmojiData
		Borders    BordersData
		ANSIEscape ANSIEscapeData
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
		ANSIEscape: ANSIEscapeData{
			Escape:          "\x1b",
			Reset:           "[0m",
			TrueColor:       "2",
			ForeGroundColor: "[38",
			BackGroundColor: "[48",
		},
	}
)

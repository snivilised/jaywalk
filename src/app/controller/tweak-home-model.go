package controller

import (
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ---------------------------------------------------------------------------
// Messages
// ---------------------------------------------------------------------------

// entrySelectedMsg is sent when the user selects a home screen entry
// by pressing 1-4. For the skeleton, entries are non-functional.
// In later issues this triggers navigation to the respective screen.
type entrySelectedMsg struct{}

// ---------------------------------------------------------------------------
// Style definitions
// ---------------------------------------------------------------------------
//
// The tweak home screen uses a fixed high-contrast style set that is
// never user-configurable. This ensures the tweak UI remains legible
// regardless of the colours the user experiments with in the embedded
// navigation preview (see design principle 3).
//
// Styles are built once at model construction time after detecting the
// terminal's light/dark background via lipgloss.HasDarkBackground.
// This prevents the home screen from becoming illegible when the user
// has a light terminal theme.

// homeStyles holds all lipgloss styles used by the tweak home screen.
type homeStyles struct {
	title       lipgloss.Style
	itemKey     lipgloss.Style
	itemName    lipgloss.Style
	itemDesc    lipgloss.Style
	footerKey   lipgloss.Style
	footerLabel lipgloss.Style
	divider     lipgloss.Style
}

// newHomeStyles builds the appropriate style set based on whether the
// terminal has a dark or light background. Called once per model.
func newHomeStyles(isDark bool) homeStyles {
	if isDark {
		return homeStyles{
			title: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#333333")).
				Padding(0, 1),

			itemKey: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")),

			itemName: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E0E0E0")),

			itemDesc: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")),

			footerKey: lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7C3AED")),

			footerLabel: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA")),

			divider: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#555555")),
		}
	}

	return homeStyles{
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1A1A1A")).
			Background(lipgloss.Color("#E8E8E8")).
			Padding(0, 1),

		itemKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")),

		itemName: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#333333")),

		itemDesc: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),

		footerKey: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")),

		footerLabel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")),

		divider: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#CCCCCC")),
	}
}

// entry describes a home screen entry point.
type entry struct {
	key  string
	name string
	desc string
}

var homeEntries = []entry{
	{key: "1", name: "Gradient Workshop", desc: "Define seed gradient, harvest steps to palette roles"},
	{key: "2", name: "Palette Editor", desc: "Edit palette roles directly"},
	{key: "3", name: "Bindings", desc: "Map component slots to gradients"},
	{key: "4", name: "Import Theme", desc: "Convert iTerm2 / VS Code / Alacritty / Warp themes"},
}

// ---------------------------------------------------------------------------
// tweakHomeModel
// ---------------------------------------------------------------------------

// TweakHomeModel is the Bubble Tea model for the tweak home screen.
// It displays the four entry points, the current theme name, and
// keyboard shortcuts for file operations, undo, and quit.
type TweakHomeModel struct {
	coordinator *TweakCoordinator
	themeName   string
	width       int
	height      int
	styles      homeStyles
}

// NewTweakHomeModel creates a TweakHomeModel bound to the given
// coordinator. The model reads state (theme name, dirty flags)
// from the coordinator but never writes to it directly.
//
// Terminal light/dark background is detected once at construction
// time so the home screen is legible regardless of the user's
// terminal theme.
func NewTweakHomeModel(tc *TweakCoordinator) TweakHomeModel {
	isDark := lipgloss.HasDarkBackground(os.Stdin, os.Stdout)

	return TweakHomeModel{
		coordinator: tc,
		themeName:   tc.themeName,
		styles:      newHomeStyles(isDark),
	}
}

func (m TweakHomeModel) Init() tea.Cmd {
	return nil
}

func (m TweakHomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case entrySelectedMsg:
		// Non-functional in skeleton. Later issues handle this
		// to navigate to the selected entry point screen.

	case tea.KeyMsg:
		switch msg.String() {
		case "1", "2", "3", "4":
			return m, nil

		case "f", "F":
			// File menu - non-functional in skeleton.

		case "z", "Z":
			m.coordinator.Undo()

		case "q", "Q":
			if m.coordinator.ExitFlow() {
				return m, tea.Quit
			}

		case "ctrl+c":
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m TweakHomeModel) View() tea.View {
	var b strings.Builder

	// Title bar
	title := fmt.Sprintf(" jay tweak        [theme: %s] ", m.themeName)
	b.WriteString(m.styles.title.Render(title))
	b.WriteString("\n\n")

	// Entry points
	for _, e := range homeEntries {
		b.WriteString("  ")
		b.WriteString(m.styles.itemKey.Render(e.key + ". "))
		b.WriteString(m.styles.itemName.Render(e.name))
		b.WriteString("\n")
		b.WriteString("       ")
		b.WriteString(m.styles.itemDesc.Render(e.desc))
		b.WriteString("\n\n")
	}

	// Footer divider
	if m.width > 0 {
		b.WriteString(m.styles.divider.Render(strings.Repeat("─", m.width)))
	} else {
		b.WriteString(m.styles.divider.Render(strings.Repeat("─", 40)))
	}
	b.WriteString("\n")

	// Footer shortcuts
	b.WriteString("  ")
	b.WriteString(m.styles.footerKey.Render("F"))
	b.WriteString(m.styles.footerLabel.Render(" File"))
	b.WriteString("    ")
	b.WriteString(m.styles.footerKey.Render("Z"))
	b.WriteString(m.styles.footerLabel.Render(" Undo"))
	b.WriteString("    ")
	b.WriteString(m.styles.footerKey.Render("Q"))
	b.WriteString(m.styles.footerLabel.Render(" Quit"))
	b.WriteString("\n")

	return tea.NewView(b.String())
}

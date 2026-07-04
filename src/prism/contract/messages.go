package contract

// ThemeUpdateMsg carries a freshly built Theme to the active view
// model. Sent by tweak whenever a working-state palette change occurs.
// Each view model's Update method gains a ThemeUpdateMsg case that
// stores the new theme. The next render uses it automatically.
type ThemeUpdateMsg struct {
	Theme Theme
}

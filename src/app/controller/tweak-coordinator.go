package controller

import (
	"context"
	"log/slog"

	tea "charm.land/bubbletea/v2"
	"github.com/snivilised/jaywalk/src/prism/contract"
)

// TweakCoordinatorOptions configures the TweakCoordinator.
type TweakCoordinatorOptions struct {
	// PreviewPath is the directory used for the live preview traversal.
	PreviewPath string

	// Palette is the initial palette loaded at startup (layer 0).
	Palette contract.Palette

	// ThemeName is the display name of the loaded theme (e.g. "starship").
	ThemeName string

	// Logger is the structured logger.
	Logger *slog.Logger
}

// tweakDirty tracks whether the user has made changes relative to the
// original loaded state.
type tweakDirty struct {
	upscale  bool // layer 1 differs from layer 0 (upscaling derived new values)
	creative bool // layer 2 differs from layer 1 (user made creative changes)
}

// TweakCoordinator manages the tweak TUI lifecycle, three-layer state
// model, dirty tracking, undo, exit flow, and the perpetual preview
// traversal auto-restart loop.
type TweakCoordinator struct {
	previewPath string

	// three-layer state model (see design doc sections 5.1-5.3)
	layer0 contract.Palette // raw loaded (read-only, never modified)
	layer1 contract.Palette // upscaled (read-only, produced from layer0)
	layer2 contract.Palette // working (mutable, what the user edits)

	dirty tweakDirty

	// program is the Bubble Tea program for the tweak home screen.
	program *tea.Program

	// cancel is the context cancellation function for the preview
	// traversal auto-restart loop.
	cancel context.CancelFunc

	themeName string
	logger    *slog.Logger
}

// NewTweakCoordinator creates a TweakCoordinator with the palette
// loaded into all three layers. Layer 1 upscaling is a no-op until
// Issue #1 implements UpscalePalette.
func NewTweakCoordinator(opts TweakCoordinatorOptions) *TweakCoordinator {
	return &TweakCoordinator{
		previewPath: opts.PreviewPath,
		layer0:      opts.Palette,
		layer1:      opts.Palette,
		layer2:      opts.Palette,
		themeName:   opts.ThemeName,
		logger:      opts.Logger,
	}
}

// Run starts the tweak TUI and blocks until the user exits.
func (tc *TweakCoordinator) Run(ctx context.Context) error {
	ctx, tc.cancel = context.WithCancel(ctx)
	defer tc.cancel()

	model := NewTweakHomeModel(tc)
	tc.program = tea.NewProgram(model)

	// Start the preview traversal auto-restart loop in the background.
	// For the skeleton, the loop body is a no-op — real traversal
	// integration is part of Issue #6.
	go tc.runPreviewLoop(ctx)

	_, err := tc.program.Run()
	return err
}

// runPreviewLoop runs agenor traversals in a perpetual loop until
// the context is cancelled. Skeleton: no-op body.
func (tc *TweakCoordinator) runPreviewLoop(ctx context.Context) {
	<-ctx.Done()
}

// Stop signals the auto-restart loop and the Bubble Tea program to
// shut down. Safe to call multiple times.
func (tc *TweakCoordinator) Stop() {
	if tc.cancel != nil {
		tc.cancel()
	}
	if tc.program != nil {
		tc.program.Quit()
	}
}

// ---------------------------------------------------------------------------
// State model operations
// ---------------------------------------------------------------------------

// WorkingPalette returns the current working state palette (layer 2).
func (tc *TweakCoordinator) WorkingPalette() contract.Palette {
	return tc.layer2
}

// Undo resets layer 2 to a copy of layer 1 and clears the creative
// dirty flag. Upscale dirty is preserved (upscaling is never lost).
func (tc *TweakCoordinator) Undo() {
	tc.layer2 = tc.layer1
	tc.dirty.creative = false
}

// IsDirty returns true when either the upscale or creative dirty flag
// is set, meaning there are unsaved changes.
func (tc *TweakCoordinator) IsDirty() bool {
	return tc.dirty.upscale || tc.dirty.creative
}

// ExitFlow determines whether the application should exit based on
// dirty state:
//   - No changes: returns true (exit silently).
//   - Changes pending: returns true (skeleton; prompt deferred to
//     later issue using huh forms).
//   - Ctrl-C: returns true (exit without saving).
func (tc *TweakCoordinator) ExitFlow() bool {
	if !tc.IsDirty() {
		return true
	}

	// Skeleton: exit flow prompts are deferred to a later issue
	// when huh forms are integrated for the confirmation dialog.
	// For now, all dirty states silently proceed to exit.
	return true
}

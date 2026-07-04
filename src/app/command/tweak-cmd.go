package command

import (
	"context"
	"os"

	"github.com/snivilised/li18ngo"
	"github.com/snivilised/mamba/assist"
	"github.com/spf13/cobra"

	"github.com/snivilised/jaywalk/src/app/controller"
	"github.com/snivilised/jaywalk/src/locale"
)

func (b *Bootstrap) buildTweakCommand(container *assist.CobraContainer) {
	tweakCmd := &cobra.Command{
		Use:   "tweak",
		Short: li18ngo.Text(locale.TweakCmdShortDescTemplData{}),
		Long:  li18ngo.Text(locale.TweakCmdLongDescTemplData{}),
		Args:  cobra.NoArgs,
		RunE:  b.runTweak,
	}

	container.MustRegisterRootedCommand(tweakCmd)
}

func (b *Bootstrap) runTweak(cmd *cobra.Command, _ []string) error {
	// Resolve the current palette from the theme loader. The palette
	// was loaded in PersistentPreRunE (via --theme flag resolution)
	// and is available through b.themeLoader.
	palette, err := b.themeLoader.Load(b.rootPs.Native.Theme)
	if err != nil {
		return err
	}

	// Resolve the preview path. For the skeleton, default to $HOME.
	// In later issues this reads from jay.ui.yml tweak.preview-path.
	previewPath := os.Getenv("HOME")
	if previewPath == "" {
		previewPath = "/"
	}

	// Create the TweakCoordinator.
	coordinator := controller.NewTweakCoordinator(
		controller.TweakCoordinatorOptions{
			PreviewPath: previewPath,
			Palette:     palette,
			ThemeName:   b.rootPs.Native.Theme,
			Logger:      b.logger,
		},
	)

	// Create a cancellable context so Ctrl-C propagates cleanly.
	ctx, cancel := context.WithCancel(cmd.Context())
	defer cancel()

	return coordinator.Run(ctx)
}

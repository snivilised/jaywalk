package command

import (
	"github.com/snivilised/li18ngo"
	"github.com/snivilised/mamba/assist"
	"github.com/spf13/cobra"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/app/controller"
	"github.com/snivilised/jaywalk/src/locale"
)

func (b *Bootstrap) buildWalkCommand(container *assist.CobraContainer) {
	walkCmd := &cobra.Command{
		Use:   "walk <directory>",
		Short: li18ngo.Text(locale.WalkCmdShortDescTemplData{}),
		Long:  li18ngo.Text(locale.WalkCmdLongDescTemplData{}),
		Args:  cobra.ExactArgs(1),
		RunE:  b.runWalk,
	}

	b.bindNavFlags(walkCmd, &b.walk.navState)

	b.walk.execPs = assist.NewParamSet[ExecParameterSet](walkCmd)
	b.bindExecFlags(walkCmd, b.walk.execPs)

	container.MustRegisterParamSet(WalkNavPsName, b.walk.navPs)
	container.MustRegisterParamSet(WalkExecPsName, b.walk.execPs)
	container.MustRegisterParamSet(WalkPreviewFamName, b.walk.previewFam)
	container.MustRegisterParamSet(WalkCascadeFamName, b.walk.cascadeFam)
	container.MustRegisterParamSet(WalkSamplingFamName, b.walk.samplingFam)
	container.MustRegisterParamSet(WalkPolyFamName, b.walk.polyFam)

	walkCmd.MarkFlagsOneRequired("action", "pipeline")
	container.MustRegisterRootedCommand(walkCmd)
}

// runWalk is the RunE handler for the walk command.
func (b *Bootstrap) runWalk(cmd *cobra.Command, args []string) error {
	subscription, err := controller.ResolveSubscription(b.walk.navPs.Native.Subscribe)
	if err != nil {
		return err
	}

	settings := controller.BuildTraversalSettings(
		createTraversalSettingsIntent(navFamilies(&b.walk.navState)),
		b.UI,
	)
	isPrime := b.walk.execPs.Native.Resume == ""

	base := controller.Request{
		Subscription: subscription,
		Settings:     settings,
		ActionName:   b.walk.navPs.Native.Action,
		PipelineName: b.walk.navPs.Native.Pipeline,
		Scenario:     agenor.Tortoise(isPrime),
		UI:           b.UI,
		GetForest:    b.options.GetForest,
		DryRun:       b.walk.previewFam.Native.DryRun,
	}

	if isPrime {
		return b.coord.ExecutePrime(cmd.Context(), &controller.PrimeRequest{
			Request: base,
			Tree:    args[0],
		})
	}

	strategy, err := resolveResumeStrategy(b.walk.execPs.Native.Resume)
	if err != nil {
		return err
	}

	return b.coord.ExecuteResume(cmd.Context(), &controller.ResumeRequest{
		Request:  base,
		Strategy: strategy,
	})
}

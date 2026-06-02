package command

import (
	"sync"

	"github.com/snivilised/li18ngo"
	"github.com/snivilised/mamba/assist"
	"github.com/snivilised/mamba/store"
	"github.com/spf13/cobra"

	"github.com/snivilised/jaywalk/src/agenor"
	"github.com/snivilised/jaywalk/src/app/controller"
	"github.com/snivilised/jaywalk/src/locale"
)

func (b *Bootstrap) buildSprintCommand(container *assist.CobraContainer) {
	sprintCmd := &cobra.Command{
		Use:   "sprint <directory>",
		Short: li18ngo.Text(locale.SprintCmdShortDescTemplData{}),
		Long:  li18ngo.Text(locale.SprintCmdLongDescTemplData{}),
		Args:  cobra.ExactArgs(1),
		RunE:  b.runSprint,
	}

	b.bindNavFlags(sprintCmd, &b.sprint.navState)

	b.sprint.execPs = assist.NewParamSet[ExecParameterSet](sprintCmd)
	b.bindExecFlags(sprintCmd, b.sprint.execPs)

	// family: worker-pool [--cpu, --now] - sprint-exclusive, local flags only.
	b.sprint.workerPoolFam = assist.NewParamSet[store.WorkerPoolParameterSet](sprintCmd)
	b.sprint.workerPoolFam.Native.BindAll(b.sprint.workerPoolFam, sprintCmd.Flags())

	container.MustRegisterParamSet(SprintNavPsName, b.sprint.navPs)
	container.MustRegisterParamSet(SprintExecPsName, b.sprint.execPs)
	container.MustRegisterParamSet(SprintPreviewFamName, b.sprint.previewFam)
	container.MustRegisterParamSet(SprintCascadeFamName, b.sprint.cascadeFam)
	container.MustRegisterParamSet(SprintSamplingFamName, b.sprint.samplingFam)
	container.MustRegisterParamSet(SprintPolyFamName, b.sprint.polyFam)
	container.MustRegisterParamSet(SprintWorkerPoolFamName, b.sprint.workerPoolFam)

	sprintCmd.MarkFlagsOneRequired("action", "pipeline")
	container.MustRegisterRootedCommand(sprintCmd)
}

// runSprint is the RunE handler for the sprint command.
func (b *Bootstrap) runSprint(cmd *cobra.Command, args []string) error {
	subscription, err := controller.ResolveSubscription(b.sprint.navPs.Native.Subscribe)
	if err != nil {
		return err
	}

	settings := controller.BuildTraversalSettings(
		createTraversalSettingsIntent(navFamilies(&b.sprint.navState)),
		b.UI,
	)

	if b.sprint.workerPoolFam.Native.CPU {
		settings = append(settings, agenor.WithCPU())
	} else if n := b.sprint.workerPoolFam.Native.NoWorkers; n > 0 {
		settings = append(settings, agenor.WithNoW(uint(n)))
	}

	isPrime := b.sprint.execPs.Native.Resume == ""
	wg := &sync.WaitGroup{}

	base := controller.Request{
		Subscription: subscription,
		Settings:     settings,
		ActionName:   b.sprint.navPs.Native.Action,
		PipelineName: b.sprint.navPs.Native.Pipeline,
		Scenario:     agenor.Hare(isPrime, wg),
		IsConcurrent: true,
		UI:           b.UI,
		GetForest:    b.options.GetForest,
		DryRun:       b.sprint.previewFam.Native.DryRun,
	}

	var execErr error

	if isPrime {
		execErr = b.coord.ExecutePrime(cmd.Context(), &controller.PrimeRequest{
			Request: base,
			Tree:    args[0],
		})
	} else {
		strategy, e := resolveResumeStrategy(b.sprint.execPs.Native.Resume)
		if e != nil {
			return e
		}

		execErr = b.coord.ExecuteResume(cmd.Context(), &controller.ResumeRequest{
			Request:  base,
			Strategy: strategy,
		})
	}

	wg.Wait()

	return execErr
}

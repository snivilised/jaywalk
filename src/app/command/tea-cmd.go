package command

import (
	"math/rand/v2"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/snivilised/mamba/assist"
	"github.com/spf13/cobra"

	"github.com/snivilised/jaywalk/src/app/bedrock"
	"github.com/snivilised/jaywalk/src/prism/traffic"
	"github.com/snivilised/jaywalk/src/prism/traffic/demo"
)

func (b *Bootstrap) buildTeaCommand(container *assist.CobraContainer) {
	teaCmd := &cobra.Command{
		Use:   "tea [directory]",
		Short: "(temporary) highway view demo — will be removed",
		Args:  cobra.MaximumNArgs(1),
		RunE:  b.runTea,
	}

	container.MustRegisterRootedCommand(teaCmd)
}

func (b *Bootstrap) runTea(_ *cobra.Command, args []string) error {
	var cfg bedrock.HighwayConfig
	if err := b.viewConfigLoader.Load("highway", &cfg); err != nil {
		return err
	}
	if cfg.Pool == "" {
		cfg.Pool = "😎 😄 👽 🤖 🦊 🐯 👻 🧙 🦄 🦁"
	}

	emojis := strings.Fields(cfg.Pool)

	deck := make([]string, len(emojis))
	copy(deck, emojis)
	rand.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})

	names := []string{"film-strip", "space-filled", "spinner", "film-strip"}
	labels := []string{"Worker-1", "Worker-2", "Worker-3", "Worker-4"}

	def, _ := traffic.Lookup(names[0])
	initialLane := demo.Lane{
		Emoji:     deck[0],
		Label:     labels[0],
		FrameFunc: def.Frames,
	}

	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	model := demo.NewModel([]demo.Lane{initialLane}, 50*time.Millisecond, root)
	p := tea.NewProgram(model)

	go func() {
		intervals := []time.Duration{3 * time.Second, 3 * time.Second, 4 * time.Second}
		deck = deck[1:]

		for i := 1; i < len(names); i++ {
			time.Sleep(intervals[i-1])

			def, ok := traffic.Lookup(names[i])
			if !ok {
				def, _ = traffic.Lookup("spinner")
			}

			pos := (i - 1) % len(deck)
			if pos == 0 && i > 1 {
				rand.Shuffle(len(deck), func(i, j int) {
					deck[i], deck[j] = deck[j], deck[i]
				})
			}
			p.Send(demo.OnboardMsg{
				Lane: demo.Lane{
					Emoji:     deck[pos],
					Label:     labels[i],
					FrameFunc: def.Frames,
				},
			})
		}
	}()

	_, err := p.Run()
	return err
}

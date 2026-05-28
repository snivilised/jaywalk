package bedrock

import (
	"errors"
	"fmt"
	"strings"

	"github.com/snivilised/jaywalk/src/prism/traffic"
)

// ValidationError collects every validation failure in one pass so the
// user sees all problems at once rather than fixing them one at a time.
type ValidationError struct {
	Failures []string
}

func (e *ValidationError) Error() string {
	return "cfg: validation failed:\n  - " + strings.Join(e.Failures, "\n  - ")
}

func (e *ValidationError) add(msg string) {
	e.Failures = append(e.Failures, msg)
}

func (e *ValidationError) addF(format string, args ...any) {
	e.add(fmt.Sprintf(format, args...))
}

func (e *ValidationError) asError() error {
	if len(e.Failures) == 0 {
		return nil
	}
	return e
}

// Validator interface - implemented per-section;
type Validator interface {
	Validate() error
}

var validLogLevels = map[string]struct{}{
	"trace": {}, "debug": {}, "info": {},
	"warn": {}, "error": {}, "fatal": {}, "panic": {},
}

// Validate checks LoggingConfig for obviously wrong values.
func (c LoggingConfig) Validate() error {
	ve := &ValidationError{}

	if c.MaxSizeInMB < 0 {
		ve.addF("logging.max-size-in-mb must be ≥ 0, got %d", c.MaxSizeInMB)
	}
	if c.MaxBackups < 0 {
		ve.addF("logging.max-backups must be ≥ 0, got %d", c.MaxBackups)
	}
	if c.MaxAgeInDays < 0 {
		ve.addF("logging.max-age-in-days must be ≥ 0, got %d", c.MaxAgeInDays)
	}
	if c.Level != "" {
		if _, ok := validLogLevels[strings.ToLower(c.Level)]; !ok {
			ve.addF("logging.level %q is not a recognised level", c.Level)
		}
	}

	return ve.asError()
}

// Validate checks AdvancedConfig.
func (c AdvancedConfig) Validate() error {
	return nil
}

// Validate checks HighwayConfig.
func (c HighwayConfig) Validate() error {
	ve := &ValidationError{}

	if c.WorkerPool != "" {
		emojis := strings.Fields(c.WorkerPool)
		if len(emojis) < 10 {
			ve.addF("ui.highway.worker-emoji-pool must have at least 10 emojis, got %d", len(emojis))
		}
	}

	if c.JobPool != "" {
		emojis := strings.Fields(c.JobPool)
		if len(emojis) < 4 {
			ve.addF("ui.highway.job-emoji-pool must have at least 4 emojis, got %d", len(emojis))
		}
	}

	spinners := &c.AnimationData.Spinners
	for _, name := range traffic.ExpandNames(spinners.Enabled) {
		cfg := spinners.Override[name]
		if _, known := traffic.SpinnerNames[name]; !known {
			ve.addF("ui.highway.animation.spinners.enabled: unknown spinner type %q", name)
			continue
		}
		if cfg != nil && cfg.Interval <= 0 {
			ve.addF("ui.highway.animation.spinners.override.%s.interval must be > 0, got %d", name, cfg.Interval)
		}
	}

	return ve.asError()
}

// Validate checks InteractionConfig.
func (c InteractionConfig) Validate() error {
	ve := &ValidationError{}
	if c.TUI.PerItemDelay < 0 {
		ve.addF("interaction.tui.per-item-delay must be ≥ 0, got %v", c.TUI.PerItemDelay)
	}
	return ve.asError()
}

// Validate checks FlagsConfig for structural consistency.
func (c FlagsConfig) Validate() error {
	ve := &ValidationError{}

	for cmd, flags := range c.Short {
		for flag, letter := range flags {
			if len(letter) != 1 {
				ve.addF(
					"flags.short.overrides.cmds.%s.%s: short letter %q must be exactly one character",
					cmd, flag, letter,
				)
			}
		}
	}

	return ve.asError()
}

// Validate checks RawAction entries.
func validateActions(actions map[string]RawAction) error {
	ve := &ValidationError{}
	for name, a := range actions {
		if strings.TrimSpace(a.Cmd) == "" {
			ve.addF("actions.%s: cmd must not be empty", name)
		}
	}
	return ve.asError()
}

// Validate checks RawPipeline entries for references to known actions.
func validatePipelines(pipelines map[string]RawPipeline, actions map[string]RawAction) error {
	ve := &ValidationError{}
	for name, p := range pipelines {
		if len(p.Steps) == 0 {
			ve.addF("pipelines.%s: steps must not be empty", name)
		}
		for i, step := range p.Steps {
			if _, ok := actions[step]; !ok {
				ve.addF("pipelines.%s.steps[%d]: references unknown action %q", name, i, step)
			}
		}
	}
	return ve.asError()
}

// Validate runs all section validators and aggregates failures.
func (c *Config) Validate() error {
	var errs []error

	if err := c.Mapped.Logging.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Mapped.Interaction.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Mapped.Advanced.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Mapped.Highway.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := c.Raw.Flags.Validate(); err != nil {
		errs = append(errs, err)
	}
	if err := validateActions(c.Raw.Actions); err != nil {
		errs = append(errs, err)
	}
	if err := validatePipelines(c.Raw.Pipelines, c.Raw.Actions); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

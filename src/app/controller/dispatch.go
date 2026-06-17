package controller

import (
	"context"
	"regexp"
	"strings"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/report"
	"github.com/snivilised/jaywalk/src/locale"
)

// handleServant dispatches to the appropriate per-node handler based on
// whether an action, pipeline, or neither is configured on the request.
// When c.poolExec is non-nil (concurrent mode), actions are submitted
// asynchronously to the worker pool and the handler returns immediately.
func (c *Coordinator) handleServant(
	ctx context.Context,
	servant core.Servant,
	req *Request,
	traversal *report.Traversal,
	peerInfoMap PeerInfoMap,
) error {
	node := servant.Node()

	var isLast bool

	if peerInfoMap != nil {
		if info, ok := peerInfoMap[node.Path]; ok {
			isLast = info.IsLast
		}
	}

	switch {
	case req.PipelineName != "":
		return c.executePipeline(ctx, node, req, isLast, traversal)

	case req.ActionName != "":
		if c.poolExec != nil {
			return c.executeActionAsync(ctx, node, req, isLast, traversal)
		}
		e := c.executeAction(ctx, node, req.ActionName, req.Root, req.DryRun)
		if e.Skipped {
			traversal.ActionsSkipped.Tick()
			req.UI.OnSkipEvent(&report.SkipEvent{
				DisplayEvent: report.DisplayEvent{
					Node:   node,
					IsLast: isLast,
					Name:   req.ActionName,
				},
				Placeholder:  e.Placeholder,
				ResolvedPath: e.ResolvedPath,
			})
			return nil
		}
		e.Event.IsLast = isLast
		req.UI.OnActionEvent(e.Event)
		return e.Event.Err

	default:
		req.UI.OnNodeEvent(&report.NeutralEvent{
			DisplayEvent: report.DisplayEvent{
				Node:   node,
				IsLast: isLast,
			},
		})
		return nil
	}
}

// executeActionAsync validates and expands the action, then submits the
// command to the async pool without waiting for completion. The output
// handler goroutine processes results and dispatches UI events.
func (c *Coordinator) executeActionAsync(
	ctx context.Context,
	node *core.Node,
	req *Request,
	isLast bool,
	traversal *report.Traversal,
) error {
	action, ok := c.config.Raw.Actions[req.ActionName]
	if !ok {
		req.UI.OnActionEvent(&report.ActionEvent{
			DisplayEvent: report.DisplayEvent{Node: node, Name: req.ActionName, IsLast: isLast},
			Err:          locale.NewActionNotFoundError(req.ActionName),
		})
		return nil
	}

	expResult := Expand(action.Cmd, req.Root, node)
	if expResult.Skipped {
		traversal.ActionsSkipped.Tick()
		req.UI.OnSkipEvent(&report.SkipEvent{
			DisplayEvent: report.DisplayEvent{
				Node:   node,
				IsLast: isLast,
				Name:   req.ActionName,
			},
			Placeholder:  expResult.Placeholder,
			ResolvedPath: expResult.ResolvedPath,
		})
		return nil
	}

	event := &report.ActionEvent{
		DisplayEvent:    report.DisplayEvent{Node: node, Name: req.ActionName},
		ExecutionString: expResult.Cmd,
		DryRun:          req.DryRun,
	}

	if req.DryRun {
		event.IsLast = isLast
		req.UI.OnActionEvent(event)
		return nil
	}

	if _, err := c.poolExec.Post(ctx, expResult.Cmd, func(workerID, _ string, output []byte, err error) {
		if err != nil {
			event.Err = err
		}
		event.CommandOutput = c.processOutput(output, c.actionRegexes[req.ActionName])
		event.WorkerID = workerID
		event.IsLast = isLast
		req.UI.OnActionEvent(event)
	}); err != nil {
		event.Err = err
		event.IsLast = isLast
		req.UI.OnActionEvent(event)
	}

	return nil
}

// actionResult is the outcome of executeAction. Either Skipped is true
// (and Placeholder/ResolvedPath carry the breach details), or Event
// carries the ActionEvent to hand to the UI.
type actionResult struct {
	Skipped      bool
	Placeholder  string
	ResolvedPath string
	Event        *report.ActionEvent
}

// executeAction expands the cmd string for the named action and returns
// the result. If a placeholder breaches root the result is marked as
// skipped and no shell execution is attempted.
func (c *Coordinator) executeAction(
	ctx context.Context,
	node *core.Node,
	name, root string,
	dryRun bool,
) actionResult {
	action, ok := c.config.Raw.Actions[name]
	if !ok {
		// PreFlight should have caught this - treat as an action error.
		return actionResult{
			Event: &report.ActionEvent{
				DisplayEvent: report.DisplayEvent{Node: node, Name: name},
				Err:          locale.NewActionNotFoundError(name),
			},
		}
	}

	result := Expand(action.Cmd, root, node)
	if result.Skipped {
		return actionResult{
			Skipped:      true,
			Placeholder:  result.Placeholder,
			ResolvedPath: result.ResolvedPath,
		}
	}

	event := &report.ActionEvent{
		DisplayEvent:    report.DisplayEvent{Node: node, Name: name},
		ExecutionString: result.Cmd,
		DryRun:          dryRun,
	}

	if !dryRun {
		output, err := c.exec(ctx, event.ExecutionString)
		if err != nil {
			event.Err = err
		}
		event.CommandOutput = c.processOutput(output, c.actionRegexes[name])
		event.WorkerID = c.lastWorkerID
	}

	return actionResult{
		Event: event,
	}
}

// executePipeline expands and executes each step in the named pipeline
// in order. The first skip or error stops the pipeline for this node.
// Each step is emitted to the UI as it completes.
func (c *Coordinator) executePipeline(ctx context.Context,
	node *core.Node,
	req *Request,
	isLast bool,
	traversal *report.Traversal,
) error {
	pipeline, ok := c.config.Raw.Pipelines[req.PipelineName]
	if !ok {
		req.UI.OnPipelineEvent(&report.PipelineEvent{
			DisplayEvent: report.DisplayEvent{
				Node:   node,
				IsLast: isLast,
				Name:   req.PipelineName,
			},
			Err: locale.NewPipelineNotFoundError(req.PipelineName),
		})
		return nil
	}

	// Emit pipeline header
	req.UI.OnPipelineEvent(&report.PipelineEvent{
		DisplayEvent: report.DisplayEvent{
			Node:             node,
			IsLast:           isLast,
			Name:             req.PipelineName,
			IsPipelineHeader: true,
		},
		DryRun: req.DryRun,
	})

	for i, step := range pipeline.Steps {
		isLastStep := i == len(pipeline.Steps)-1
		ar := c.executeAction(ctx, node, step, req.Root, req.DryRun)

		if ar.Skipped {
			traversal.ActionsSkipped.Tick()
			req.UI.OnSkipEvent(&report.SkipEvent{
				DisplayEvent: report.DisplayEvent{
					Node:           node,
					IsLast:         isLast,
					Name:           step,
					IsPipelineStep: true,
					IsLastStep:     true, // skip always terminates
				},
				Placeholder:  ar.Placeholder,
				ResolvedPath: ar.ResolvedPath,
			})
			return nil
		}

		ar.Event.IsLast = isLast
		ar.Event.IsPipelineStep = true
		ar.Event.IsLastStep = isLastStep && ar.Event.Err == nil // error terminates but we show it as a step

		req.UI.OnActionEvent(ar.Event)

		if ar.Event.Err != nil {
			return ar.Event.Err
		}
	}

	return nil
}

const (
	minLimit     = 20
	maxLimit     = 120
	defaultLimit = 75
	ellipsis     = " ..."
)

// processOutput extracts a single line from the raw command output and applies truncation.
// It removes leading/trailing empty lines. If captureRe is provided, it uses it
// to select the matching line.
func (c *Coordinator) processOutput(output []byte, captureRe *regexp.Regexp) string {
	lines := strings.Split(string(output), "\n")
	var contentLines []string
	for _, ln := range lines {
		trimmed := strings.TrimSpace(ln)
		if trimmed != "" {
			contentLines = append(contentLines, trimmed)
		}
	}

	if len(contentLines) == 0 {
		return ""
	}

	selectedLine := contentLines[0]

	if captureRe != nil {
		for _, ln := range contentLines {
			if captureRe.MatchString(ln) {
				selectedLine = ln
				break
			}
		}
	}

	limit := c.config.Mapped.Advanced.Output.Exec.Truncate
	if limit < minLimit || limit > maxLimit {
		limit = defaultLimit
	}

	if len(selectedLine) > limit {
		if limit > 4 {
			selectedLine = selectedLine[:limit-len(ellipsis)] + ellipsis
		} else {
			selectedLine = selectedLine[:limit]
		}
	}

	return selectedLine
}

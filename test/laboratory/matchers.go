package lab

import (
	"fmt"
	"io/fs"
	"slices"
	"strings"

	. "github.com/onsi/gomega/types" //nolint:staticcheck // ok
	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/third/lo"
)

// DirectoryContentsMatcher is a custom Gomega matcher that checks if the
// contents of a directory match the expected list of file or directory names.
// It compares the actual names of the entries in the directory with the expected
// names and provides detailed failure messages if they do not match.
type DirectoryContentsMatcher struct {
	expected      any
	expectedNames []string
	actualNames   []string
}

// HaveDirectoryContents is a custom Gomega matcher that checks if a specific node has
// been invoked during a test.
func HaveDirectoryContents(expected any) GomegaMatcher {
	return &DirectoryContentsMatcher{
		expected: expected,
	}
}

// Match checks if the actual directory contents match the expected list of names.
func (m *DirectoryContentsMatcher) Match(actual any) (bool, error) {
	entries, entriesOk := actual.([]fs.DirEntry)
	if !entriesOk {
		return false, fmt.Errorf("🔥 matcher expected []fs.DirEntry (%T)", entries)
	}

	m.actualNames = lo.Map(entries, func(entry fs.DirEntry, _ int) string {
		return entry.Name()
	})

	expected, expectedOk := m.expected.([]string)
	if !expectedOk {
		return false, fmt.Errorf("🔥 matcher expected []string (%T)", expected)
	}

	m.expectedNames = expected

	return slices.Compare(m.actualNames, m.expectedNames) == 0, nil
}

// FailureMessage returns a detailed message when the actual directory contents
// do not match the expected names.
func (m *DirectoryContentsMatcher) FailureMessage(_ any) string {
	return fmt.Sprintf(
		"❌ DirectoryContentsMatcher Expected\n\t%v\nto match contents\n\t%v\n",
		strings.Join(m.expectedNames, ", "), strings.Join(m.actualNames, ", "),
	)
}

// NegatedFailureMessage returns a detailed message when the actual directory contents
// match the expected names, but they were expected not to.
func (m *DirectoryContentsMatcher) NegatedFailureMessage(_ any) string {
	return fmt.Sprintf(
		"❌ DirectoryContentsMatcher Expected\n\t%v\nNOT to match contents\n\t%v\n",
		strings.Join(m.expectedNames, ", "), strings.Join(m.actualNames, ", "),
	)
}

// InvokeNodeMatcher is a custom Gomega matcher that checks if a specific node has been
// invoked during a test. It verifies if the node's name exists in the recall map, which
// tracks the nodes that have been invoked. The matcher provides detailed failure messages
// if the expected node was not invoked.
type InvokeNodeMatcher struct {
	expected  any
	mandatory string
}

// HaveInvokedNode creates a new instance of InvokeNodeMatcher with the
// provided expected node name.
func HaveInvokedNode(expected any) GomegaMatcher {
	return &InvokeNodeMatcher{
		expected: expected,
	}
}

// Match checks if the expected node name exists in the recall map, indicating
// that it was invoked.
func (m *InvokeNodeMatcher) Match(actual any) (bool, error) {
	recall, ok := actual.(Recall)
	if !ok {
		return false, fmt.Errorf(
			"InvokeNodeMatcher expected actual to be a RecordingMap (%T)",
			actual,
		)
	}

	mandatory, ok := m.expected.(string)
	if !ok {
		return false, fmt.Errorf("InvokeNodeMatcher expected string (%T)", actual)
	}

	m.mandatory = mandatory

	_, found := recall[m.mandatory]

	return found, nil
}

// FailureMessage returns a detailed message when the expected node was not invoked.
func (m *InvokeNodeMatcher) FailureMessage(_ any) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode to be invoked\n",
		m.mandatory,
	)
}

// NegatedFailureMessage returns a detailed message when the expected node was invoked,
// but it was expected not to be.
func (m *InvokeNodeMatcher) NegatedFailureMessage(_ any) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode NOT to be invoked\n",
		m.mandatory,
	)
}

// NotInvokeNodeMatcher is a custom Gomega matcher that checks if a specific node has NOT been
// invoked during a test. It verifies if the node's name does NOT exist in the recall map,
// which tracks the nodes that have been invoked. The matcher provides detailed failure
// messages if the expected node was invoked when it was expected not to be.
type NotInvokeNodeMatcher struct {
	expected  any
	mandatory string
}

// HaveNotInvokedNode creates a new instance of NotInvokeNodeMatcher with the
// provided expected node name.
func HaveNotInvokedNode(expected any) GomegaMatcher {
	return &NotInvokeNodeMatcher{
		expected: expected,
	}
}

// Match checks if the expected node name does NOT exist in the recall map, indicating
// that it was not invoked.
func (m *NotInvokeNodeMatcher) Match(actual any) (bool, error) {
	recall, ok := actual.(Recall)
	if !ok {
		return false, fmt.Errorf("matcher expected actual to be a RecordingMap (%T)", actual)
	}

	mandatory, ok := m.expected.(string)
	if !ok {
		return false, fmt.Errorf("matcher expected string (%T)", actual)
	}

	m.mandatory = mandatory

	_, found := recall[m.mandatory]

	return !found, nil
}

// FailureMessage returns a detailed message when the expected node was invoked, but it
// was expected not to be.
func (m *NotInvokeNodeMatcher) FailureMessage(_ any) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode to NOT be invoked\n",
		m.mandatory,
	)
}

// NegatedFailureMessage returns a detailed message when the expected node was not invoked,
// but it was expected to be.
func (m *NotInvokeNodeMatcher) NegatedFailureMessage(_ any) string {
	return fmt.Sprintf("❌ Expected\n\t%v\nnode to be invoked\n",
		m.mandatory,
	)
}

type (
	// ExpectedCount is a struct that represents the expected count of child nodes
	// for a specific node in a test. It contains the name of the node and the
	// expected count of its child nodes.
	ExpectedCount struct {
		// Name is the name of the node for which the child count is being tested.
		Name string

		// Count is the expected number of child nodes for the specified node.
		Count core.MetricValue
	}

	// ChildCountMatcher is a custom Gomega matcher that checks if the actual
	// count of child nodes for a specific node matches the expected count.
	// It compares the actual count of child nodes retrieved from the recall
	// map with the expected count provided in the ExpectedCount struct. The matcher
	// provides detailed failure messages if the actual count does not match the
	// expected count.
	ChildCountMatcher struct {
		expected    any
		expectation MatcherExpectation[core.MetricValue]
		name        string
	}
)

// HaveChildCountOf creates a new instance of ChildCountMatcher with the provided
// expected count of child nodes for a specific node.
func HaveChildCountOf(expected any) GomegaMatcher {
	return &ChildCountMatcher{
		expected: expected,
	}
}

// Match checks if the actual count of child nodes for the specified node matches
// the expected count.
func (m *ChildCountMatcher) Match(actual any) (bool, error) {
	recall, ok := actual.(Recall)
	if !ok {
		return false, fmt.Errorf("ChildCountMatcher expected actual to be a RecordingMap (%T)", actual)
	}

	expected, ok := m.expected.(ExpectedCount)
	if !ok {
		return false, fmt.Errorf("ChildCountMatcher expected ExpectedCount (%T)", actual)
	}

	count, ok := recall[expected.Name]
	if !ok {
		return false, fmt.Errorf("🔥 not found: '%v'", expected.Name)
	}

	m.expectation = MatcherExpectation[core.MetricValue]{
		Expected: expected.Count,
		Actual:   count,
	}
	m.name = expected.Name

	return m.expectation.IsEqual(), nil
}

// FailureMessage returns a detailed message when the actual count of child
// nodes does not match the expected count for the specified node.
func (m *ChildCountMatcher) FailureMessage(_ any) string {
	return fmt.Sprintf(
		"❌ Expected child count for node: '%v' to be equal; expected: '%v', actual: '%v'\n",
		m.name, m.expectation.Expected, m.expectation.Actual,
	)
}

// NegatedFailureMessage returns a detailed message when the actual count of child
// nodes matches the expected count for the specified node, but it was expected
// not to.
func (m *ChildCountMatcher) NegatedFailureMessage(_ any) string {
	return fmt.Sprintf(
		"❌ Expected child count for node: '%v' NOT to be equal; expected: '%v', actual: '%v'\n",
		m.name, m.expectation.Expected, m.expectation.Actual,
	)
}

type (
	// ExpectedMetric is a struct that represents the expected count of a
	// specific metric in a test. It contains the type of the metric and the
	// expected count for that metric.
	ExpectedMetric struct {
		// Type is the type of the metric for which the count is being tested.
		Type enums.Metric

		// Count is the expected count for the specified metric type.
		Count core.MetricValue
	}

	// MetricMatcher is a custom Gomega matcher that checks if the actual count of a
	// specific metric matches the expected count. It compares the actual count of
	// the specified metric retrieved from the TraverseResult with the expected
	// count provided in the ExpectedMetric struct. The matcher provides detailed
	// failure messages if the actual count does not match the expected count.
	MetricMatcher struct {
		expected    any
		expectation MatcherExpectation[core.MetricValue]
		typ         enums.Metric
	}
)

// HaveMetricCountOf creates a new instance of MetricMatcher with the provided
// expected count for a specific metric type.
func HaveMetricCountOf(expected any) GomegaMatcher {
	return &MetricMatcher{
		expected: expected,
	}
}

// Match checks if the actual count of the specified metric type matches the
// expected count.
func (m *MetricMatcher) Match(actual any) (bool, error) {
	result, ok := actual.(core.TraverseResult)
	if !ok {
		return false, fmt.Errorf(
			"🔥 MetricMatcher expected actual to be a core.TraverseResult (%T)",
			actual,
		)
	}

	expected, ok := m.expected.(ExpectedMetric)
	if !ok {
		return false, fmt.Errorf("🔥 MetricMatcher expected ExpectedMetric (%T)", actual)
	}

	actualCount := result.Metrics().Count(expected.Type)
	m.expectation = MatcherExpectation[core.MetricValue]{
		Expected: expected.Count,
		Actual:   actualCount,
	}
	m.typ = expected.Type

	return m.expectation.IsEqual(), nil
}

// FailureMessage returns a detailed message when the actual count of
// the specified metric type does not match the expected count.
func (m *MetricMatcher) FailureMessage(_ any) string {
	return fmt.Sprintf(
		"❌ Expected metric '%v' to be equal; expected:'%v', actual: '%v'\n",
		m.typ.String(), m.expectation.Expected, m.expectation.Actual,
	)
}

// NegatedFailureMessage returns a detailed message when the actual
// count of the specified metric type matches the expected count, but it
// was expected not to.
func (m *MetricMatcher) NegatedFailureMessage(_ any) string {
	return fmt.Sprintf(
		"❌ Expected metric '%v' NOT to be equal; expected:'%v', actual: '%v'\n",
		m.typ.String(), m.expectation.Expected, m.expectation.Actual,
	)
}

package lab

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/snivilised/jaywalk/src/agenor/pref"
	"github.com/snivilised/jaywalk/src/agenor/test/hanno"
)

type (
	// NamedFunc is a function type that takes a name as input and returns a
	// string, typically used for generating error messages or reasons based
	// on the name of a node.
	NamedFunc func(name string) string
)

var (
	// Reasons is a struct that contains functions for generating error messages
	// or reasons based on the name of a node. It provides a convenient way to
	// create consistent error messages throughout the codebase.
	Reasons = struct {
		Node NamedFunc
	}{
		Node: func(name string) string {
			return fmt.Sprintf("❌ for node named: '%v'", name)
		},
	}
)

// Normalise is a utility function that takes a path string and replaces all
// occurrences of the forward slash ("/") with the system's file separator.
// This is useful for ensuring that paths are consistent across different
// operating systems, as Windows uses a backslash ("\") while Unix-based systems
// use a forward slash ("/").
func Normalise(p string) string {
	return strings.ReplaceAll(p, "/", string(filepath.Separator))
}

// Because is a utility function that generates an error message based on the
// name of a node and a reason for the error. It takes two string parameters: the
// name of the node and the reason for the error, and returns a formatted
// string that includes both pieces of information. This function is useful for
// creating consistent and informative error messages throughout the codebase.
func Because(name, because string) string {
	return fmt.Sprintf("❌ for node named: '%v', because: '%v'", name, because)
}

// Reason is a utility function that generates an error message based on the name
// of a node. It takes a single string parameter, the name of the node, and returns
// a formatted string that includes the name. This function is useful for creating
// consistent error messages that indicate which node is causing an issue.
func Reason(name string) string {
	return fmt.Sprintf("❌ for node named: '%v'", name)
}

// yoke is similar to filepath.Join but it is meant specifically for relative file
// systems where the rules of a path are different; see fs.ValidPath
func yoke(segments ...string) string {
	return strings.Join(segments, "/")
}

// GetJSONPath is a utility function that constructs the path to a JSON file used for
// testing. It uses the hanno.Repo function to get the base path to the "test/jason"
// directory and then appends the Static.JSONFile name to it using the yoke function.
// This function is useful for ensuring that the correct path to the JSON file is
// consistently used throughout the tests.
func GetJSONPath() string {
	jroot := hanno.Repo(filepath.Join("test", "json"))

	return yoke(jroot, "unmarshal", Static.JSONFile)
}

// GetJSONDir is a utility function that returns the path to the "test/jason" directory
// using the hanno.Repo function. This function is useful for ensuring that the correct
// path to the JSON directory is consistently used throughout the tests.
func GetJSONDir() string {
	return hanno.Repo(filepath.Join("test", "json"))
}

// IgnoreFault is a utility function that takes a pointer to a pref.NavigationFault
// and returns nil. This function can be used as a placeholder or a no-op in situations
// where a navigation fault is expected but should not cause the test to fail. By
// returning nil, it effectively ignores any navigation faults that occur during the
// test execution.
func IgnoreFault(_ *pref.NavigationFault) error {
	return nil
}

// stripANSI removes ANSI escape sequences from a string for test assertions.
func StripANSI(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}

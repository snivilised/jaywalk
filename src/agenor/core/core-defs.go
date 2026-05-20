// Package core contains universal definitions and handles user facing cross
// cutting concerns.
package core

import (
	"io/fs"
	"os"
	"time"

	"github.com/snivilised/jaywalk/src/agenor/enums"
	"github.com/snivilised/jaywalk/src/agenor/tfs"
)

type (
	// ResultCompletion used to determine if the result really represents
	// final navigation completion.
	ResultCompletion interface {
		// IsComplete returns true if the traversal has completed successfully
		// and false otherwise.
		IsComplete() bool
	}

	// Completion is a function type that can be used to determine if a traversal
	// has completed successfully. It returns true if the traversal is complete
	// and false otherwise.
	Completion func() bool

	// Session represents a traversal session and keeps tracks of
	// timing.
	Session interface {
		ResultCompletion

		// StartedAt returns the time when the traversal session started.
		StartedAt() time.Time

		// Elapsed returns the duration of the traversal session.
		Elapsed() time.Duration
	}

	// TraverseResult represents the result of a traversal
	TraverseResult interface {
		// Metrics returns the metrics collected during the traversal. This allows
		// the client to access performance data and other relevant information
		// about the traversal.
		Metrics() Reporter

		// Session returns the session information for the traversal, including
		// timing and completion status. This allows the client to access details
		// about the traversal session and determine if it completed successfully.
		Session() Session
	}

	// Servant provides the client with facility to request properties
	// about the current navigation node.
	Servant interface {
		// Node returns the current node being processed during traversal. This allows
		// the client to access information about the file system entity (file or directory)
		// that is currently being handled by the traversal process.
		Node() *Node

		// Peer returns the peer-relative position data for the current node, if available.
		// This allows the client to access information about the node's position relative
		// to its siblings and ancestors, which can be useful for rendering or other
		// processing that depends on the node's context within the file system hierarchy.
		Peer() *PeerInfo
	}

	// PeerInfo contains peer-relative position data for a node,
	// resolved after filtering.
	PeerInfo struct {
		// IsLast is true if no filtered siblings follow this node.
		IsLast bool
	}

	// Forest contains the logical file systems required
	// for navigation.
	Forest struct {
		// T is the file system that contains just the functionality required
		// for traversal. It can also represent other file systems including afero,
		// providing the appropriate adapters are in place.
		T tfs.TraversalFS

		// R is the file system required for resume operations, ie we load
		// and save resume state via this file system instance, which is
		// distinct from the traversal file system.
		R tfs.TraversalFS
	}

	// Client is the callback invoked for each file system node found
	// during traversal.
	Client func(servant Servant) error

	// FsDescription description of a file system
	FsDescription struct {
		// IsRelative indicates whether the file system is relative or absolute. This
		// can be used to determine how to interpret paths and perform operations on
		// the file system during traversal.
		IsRelative bool
	}

	// Permissions used to hold default permissions
	Permissions struct {
		// File represents file mode permission.
		File fs.FileMode

		// Dir represents directory mode permission.
		Dir fs.FileMode
	}

	// ActiveState represents state that needs to be persisted alongside
	// the options in order for resume to work.
	ActiveState struct {
		// Tree represents the root of the traversal.
		Tree string

		// TraverseDescription provides a description of the file system being traversed.
		TraverseDescription FsDescription

		// ResumeDescription provides a description of the file system used for resume operations.
		ResumeDescription FsDescription

		// Subscription represents the subscription of the traversal.
		Subscription enums.Subscription

		// Hibernation represents the hibernation state of the traversal.
		Hibernation enums.Hibernation

		// CurrentPath represents the current path being processed during traversal. This allows
		// the client to track the progress of the traversal and determine where it is in the
		// file system hierarchy.
		CurrentPath string

		// IsDir indicates whether the current path is a directory or a file. This can be used
		// to determine how to handle the current node during traversal, as directories and files
		// may require different processing.
		IsDir bool

		// Depth represents the current depth of the traversal in the file system hierarchy. This can be used
		// to track how deep the traversal has gone and can be useful for implementing depth-based
		// logic or optimizations during the traversal process.
		Depth TraversalDepth

		// Metrics represents the metrics collected during the traversal. This allows the client to access
		// performance data and other relevant information about the traversal, which can be useful for
		// monitoring and optimizing the traversal process.
		Metrics Metrics
	}

	// TimeFunc get time
	TimeFunc func() time.Time

	// HomeFunc returns the home directory path for the current user.
	HomeFunc func() (string, error)

	// GetenvFunc is a function type that takes a key as input and returns the
	// corresponding environment variable value as a string. This allows for flexible
	// retrieval of environment variables.
	GetenvFunc func(key string) string

	// SimpleHandler is a function that takes no parameters and can
	// be used by any notification with this signature.
	SimpleHandler func()

	// BeginHandler invoked before traversal begins
	BeginHandler func(tree string)

	// EndHandler invoked at the end of traversal
	EndHandler func(result TraverseResult)

	// HibernateHandler is a generic handler that is used by hibernation
	// to indicate wake or sleep.
	HibernateHandler func(description string)
)

// IsComplete returns true if the traversal has completed successfully and false otherwise.
func (fn Completion) IsComplete() bool {
	return fn()
}

// Clone creates a copy of the ActiveState instance. This can be useful for creating
// a new instance with the same state without modifying the original, allowing for
// safe concurrent access or preserving the original state for future reference.
func (s *ActiveState) Clone() *ActiveState {
	c := *s
	return &c
}

const (
	// FileSystemTimeFormat the format of the string timestamp encoded
	// into the resume file. This is a fixed format in order to enable
	// easy processing of resume files.
	FileSystemTimeFormat = "2006-01-02_15-04-05"

	// PackageName is the name of the package, used for constructing paths to
	// store administrative files and directories.
	PackageName = "agenor"

	filePerm = 0o644
	dirPerm  = 0o755
)

var (
	// Now is the function used to compute the current time
	Now TimeFunc

	// Home is the function used to compute the home directory path for the current user.
	Home HomeFunc

	// Getenv is the function used to retrieve environment variable values.
	Getenv GetenvFunc

	// Perms defines the default permissions used to create administrative
	// files and directories.
	Perms Permissions

	// ResumeTail is the trailing part of a directory location used to
	// store admin files; eg resume restoration files.
	ResumeTail string
)

func init() {
	Now = time.Now
	Home = os.UserHomeDir
	Getenv = os.Getenv

	Perms = Permissions{
		File: filePerm,
		Dir:  dirPerm,
	}
	ResumeTail = "resume"
}

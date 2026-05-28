// Package contract holds all shared type definitions for the prism
// presentation layer. It is a leaf package: both prism root and its
// child packages (flow, highway) import contract, never the reverse.
// This ensures parent packages depend on children, not the other way around.
package contract

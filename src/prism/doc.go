// Package prism defines the presentation layer used by jay, assisted by the charm
// universe. prism is also designed to be used by third parties.
//
// Architecture: shared types (Palette, Theme, Renderer, Motif, etc.) live in the
// sub-package prism/contract. Both prism root and its child packages (flow, highway)
// import contract, ensuring parent -> child dependency direction.
package prism

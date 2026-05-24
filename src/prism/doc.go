// package prism defines presentation layer used by jay, assisted by the charm
// universe. prism is also designed to be used by third parties.
package prism

// TODO: we need a refactor inside prism. There are too many instances where
// child packages are dependent on types in parent packages. That is an absurd package
// relationship. Parents can depend on children, not the other way around.
// EG highway.lane.go depends on types in prism. This suggests that types need
// to be moved around.

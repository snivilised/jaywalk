package porthole

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScroll(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scroll Suite")
}

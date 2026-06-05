package scrollbar_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestScrollbar(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Scrollbar Suite")
}

package workshop_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkshop(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workshop Suite")
}

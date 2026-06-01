package sampler

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSampler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sampler Suite")
}

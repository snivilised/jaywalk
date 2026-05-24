package traffic

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTraffic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Traffic Suite")
}

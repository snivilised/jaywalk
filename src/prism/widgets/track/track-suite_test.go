package track

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTrack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Track Suite")
}

package controller

import (
	"context"
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/snivilised/jaywalk/src/agenor/core"
	"github.com/snivilised/jaywalk/src/app/bedrock"
)

var _ = Describe("Dispatch Output Processing", func() {
	var c *Coordinator

	BeforeEach(func() {
		cfg := &bedrock.Config{}
		cfg.Mapped.Advanced.Output.Exec.Truncate = 75
		c = &Coordinator{config: cfg}
	})

	Describe("processOutput", func() {
		It("should discard leading empty lines and pick the first non-empty line", func() {
			output := []byte("\n\n  \nfirst line\nsecond line")
			result := c.processOutput(output, nil)
			Expect(result).To(Equal("first line"))
		})

		It("should return empty string if only empty lines", func() {
			output := []byte("\n\n  \n")
			result := c.processOutput(output, nil)
			Expect(result).To(Equal(""))
		})

		It("should select the line matching the capture regex", func() {
			output := []byte("first line\nsecond line with magic\nthird line")
			result := c.processOutput(output, regexp.MustCompile("magic"))
			Expect(result).To(Equal("second line with magic"))
		})

		It("should fallback to first non-empty line if capture regex finds no match", func() {
			output := []byte("first line\nsecond line\nthird line")
			result := c.processOutput(output, regexp.MustCompile("magic"))
			Expect(result).To(Equal("first line"))
		})

		It("should fallback to first non-empty line if capture regex is invalid (simulated by nil)", func() {
			output := []byte("first line\nsecond line\nthird line")
			result := c.processOutput(output, nil)
			Expect(result).To(Equal("first line"))
		})

		It("should truncate output if it exceeds limits and append ellipsis", func() {
			c.config.Mapped.Advanced.Output.Exec.Truncate = 20
			output := []byte("this is a very long line that should be truncated")
			result := c.processOutput(output, nil)
			Expect(result).To(Equal("this is a very l ..."))
			Expect(len(result)).To(Equal(20))
		})

		It("should use default truncation limit if configured value is out of bounds", func() {
			c.config.Mapped.Advanced.Output.Exec.Truncate = 10 // Invalid, min is 20
			output := []byte("this is a very long line that should be truncated normally, but let us test bounds. this string is exactly 84 characters.")
			result := c.processOutput(output, nil)
			Expect(result).To(Equal("this is a very long line that should be truncated normally, but let us  ..."))
			Expect(len(result)).To(Equal(75))
		})
	})
})

var _ = Describe("executeAction", func() {
	It("propagates context cancellation from exec", func() {
		cfg := &bedrock.Config{}
		cfg.Raw.Actions = map[string]bedrock.RawAction{
			"test-action": {Cmd: "echo hello"},
		}
		c := &Coordinator{
			config: cfg,
			exec: func(_ context.Context, _ string) ([]byte, error) {
				return nil, context.Canceled
			},
		}

		node := core.New("test-path", nil, nil, nil, nil)
		result := c.executeAction(context.Background(), node, "test-action", "", false)
		Expect(result.Event.Err).To(MatchError(context.Canceled))
	})

	It("returns nil error when exec succeeds", func() {
		cfg := &bedrock.Config{}
		cfg.Raw.Actions = map[string]bedrock.RawAction{
			"test-action": {Cmd: "echo hello"},
		}
		c := &Coordinator{
			config: cfg,
			exec: func(_ context.Context, _ string) ([]byte, error) {
				return []byte("ok"), nil
			},
		}

		node := core.New("test-path", nil, nil, nil, nil)
		result := c.executeAction(context.Background(), node, "test-action", "", false)
		Expect(result.Event.Err).To(BeNil())
	})
})

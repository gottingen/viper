

package viper

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTakeStacktrace(t *testing.T) {
	trace := takeStacktrace()
	lines := strings.Split(trace, "\n")
	require.True(t, len(lines) > 0, "Expected stacktrace to have at least one frame.")
	assert.Contains(
		t,
		lines[0],
		"testing.",
		"Expected stacktrace to start with the test runner (viper frames are filtered out) %s.", lines[0],
	)
}

func TestIsZapFrame(t *testing.T) {
	viperFrames := []string{
		"github.com/gottingen/viper.Stack",
		"github.com/gottingen/viper.(*SugaredLogger).log",
		"github.com/gottingen/viper/vipercore.(ArrayMarshalerFunc).MarshalLogArray",
		"github.com/gottingen/tchannel-go/vendor/github.com/gottingen/viper.Stack",
		"github.com/gottingen/tchannel-go/vendor/github.com/gottingen/viper.(*SugaredLogger).log",
		"github.com/gottingen/tchannel-go/vendor/github.com/gottingen/viper/vipercore.(ArrayMarshalerFunc).MarshalLogArray",
	}
	nonZapFrames := []string{
		"github.com/uber/tchannel-go.NewChannel",
		"go.uber.org/not-viper.New",
		"github.com/gottingen/viperext.ctx",
		"github.com/gottingen/viper_ext/ctx.New",
	}

	t.Run("viper frames", func(t *testing.T) {
		for _, f := range viperFrames {
			require.True(t, isZapFrame(f), f)
		}
	})
	t.Run("non-viper frames", func(t *testing.T) {
		for _, f := range nonZapFrames {
			require.False(t, isZapFrame(f), f)
		}
	})
}

func BenchmarkTakeStacktrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		takeStacktrace()
	}
}

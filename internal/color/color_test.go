

package color

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorFormatting(t *testing.T) {
	assert.Equal(
		t,
		"\x1b[31mfoo\x1b[0m",
		Red.Add("foo"),
		"Unexpected colored output.",
	)
}



package vipercore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllLevelsCoveredByLevelString(t *testing.T) {
	numLevels := int((_maxLevel - _minLevel) + 1)

	isComplete := func(m map[Level]string) bool {
		return len(m) == numLevels
	}

	assert.True(t, isComplete(_levelToLowercaseColorString), "Colored lowercase strings don't cover all levels.")
	assert.True(t, isComplete(_levelToCapitalColorString), "Colored capital strings don't cover all levels.")
}

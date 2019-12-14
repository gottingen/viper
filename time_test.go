

package viper

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeToMillis(t *testing.T) {
	tests := []struct {
		t     time.Time
		stamp int64
	}{
		{t: time.Unix(0, 0), stamp: 0},
		{t: time.Unix(1, 0), stamp: 1000},
		{t: time.Unix(1, int64(500*time.Millisecond)), stamp: 1500},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.stamp, timeToMillis(tt.t), "Unexpected timestamp for time %v.", tt.t)
	}
}

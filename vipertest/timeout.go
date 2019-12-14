

package vipertest

import (
	"time"

	"github.com/gottingen/viper/internal/vtest"
)

// Timeout scales the provided duration by $TEST_TIMEOUT_SCALE.
//
// Deprecated: This function is intended for internal testing and shouldn't be
// used outside viper itself. It was introduced before Go supported internal
// packages.
func Timeout(base time.Duration) time.Duration {
	return vtest.Timeout(base)
}

// Sleep scales the sleep duration by $TEST_TIMEOUT_SCALE.
//
// Deprecated: This function is intended for internal testing and shouldn't be
// used outside viper itself. It was introduced before Go supported internal
// packages.
func Sleep(base time.Duration) {
	vtest.Sleep(base)
}

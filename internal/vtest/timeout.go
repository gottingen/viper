

package vtest

import (
	"log"
	"os"
	"strconv"
	"time"
)

var _timeoutScale = 1.0

// Timeout scales the provided duration by $TEST_TIMEOUT_SCALE.
func Timeout(base time.Duration) time.Duration {
	return time.Duration(float64(base) * _timeoutScale)
}

// Sleep scales the sleep duration by $TEST_TIMEOUT_SCALE.
func Sleep(base time.Duration) {
	time.Sleep(Timeout(base))
}

// Initialize checks the environment and alters the timeout scale accordingly.
// It returns a function to undo the scaling.
func Initialize(factor string) func() {
	original := _timeoutScale
	fv, err := strconv.ParseFloat(factor, 64)
	if err != nil {
		panic(err)
	}
	_timeoutScale = fv
	return func() { _timeoutScale = original }
}

func init() {
	if v := os.Getenv("TEST_TIMEOUT_SCALE"); v != "" {
		Initialize(v)
		log.Printf("Scaling timeouts by %vx.\n", _timeoutScale)
	}
}

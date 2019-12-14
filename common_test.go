

package viper

import (
	"sync"
	"testing"

	"github.com/gottingen/viper/vipercore"
	"github.com/gottingen/viper/vipertest/observer"
)

func opts(opts ...Option) []Option {
	return opts
}

// Here specifically to introduce an easily-identifiable filename for testing
// stacktraces and caller skips.
func withLogger(t testing.TB, e vipercore.LevelEnabler, opts []Option, f func(*Logger, *observer.ObservedLogs)) {
	fac, logs := observer.New(e)
	log := New(fac, opts...)
	f(log, logs)
}

func withSugar(t testing.TB, e vipercore.LevelEnabler, opts []Option, f func(*SugaredLogger, *observer.ObservedLogs)) {
	withLogger(t, e, opts, func(logger *Logger, logs *observer.ObservedLogs) { f(logger.Sugar(), logs) })
}

func runConcurrently(goroutines, iterations int, wg *sync.WaitGroup, f func()) {
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				f()
			}
		}()
	}
}

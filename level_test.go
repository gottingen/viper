

package viper

import (
	"sync"
	"testing"

	"github.com/gottingen/viper/vipercore"

	"github.com/stretchr/testify/assert"
)

func TestLevelEnablerFunc(t *testing.T) {
	enab := LevelEnablerFunc(func(l vipercore.Level) bool { return l == vipercore.InfoLevel })
	tests := []struct {
		level   vipercore.Level
		enabled bool
	}{
		{DebugLevel, false},
		{InfoLevel, true},
		{WarnLevel, false},
		{ErrorLevel, false},
		{DPanicLevel, false},
		{PanicLevel, false},
		{FatalLevel, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.enabled, enab.Enabled(tt.level), "Unexpected result applying LevelEnablerFunc to %s", tt.level)
	}
}

func TestNewAtomicLevel(t *testing.T) {
	lvl := NewAtomicLevel()
	assert.Equal(t, InfoLevel, lvl.Level(), "Unexpected initial level.")
	lvl.SetLevel(ErrorLevel)
	assert.Equal(t, ErrorLevel, lvl.Level(), "Unexpected level after SetLevel.")
	lvl = NewAtomicLevelAt(WarnLevel)
	assert.Equal(t, WarnLevel, lvl.Level(), "Unexpected level after SetLevel.")
}

func TestAtomicLevelMutation(t *testing.T) {
	lvl := NewAtomicLevel()
	lvl.SetLevel(WarnLevel)
	// Trigger races for non-atomic level mutations.
	proceed := make(chan struct{})
	wg := &sync.WaitGroup{}
	runConcurrently(10, 100, wg, func() {
		<-proceed
		assert.Equal(t, WarnLevel, lvl.Level())
	})
	runConcurrently(10, 100, wg, func() {
		<-proceed
		lvl.SetLevel(WarnLevel)
	})
	close(proceed)
	wg.Wait()
}

func TestAtomicLevelText(t *testing.T) {
	tests := []struct {
		text   string
		expect vipercore.Level
		err    bool
	}{
		{"debug", DebugLevel, false},
		{"info", InfoLevel, false},
		{"", InfoLevel, false},
		{"warn", WarnLevel, false},
		{"error", ErrorLevel, false},
		{"dpanic", DPanicLevel, false},
		{"panic", PanicLevel, false},
		{"fatal", FatalLevel, false},
		{"foobar", InfoLevel, true},
	}

	for _, tt := range tests {
		var lvl AtomicLevel
		// Test both initial unmarshaling and overwriting existing value.
		for i := 0; i < 2; i++ {
			if tt.err {
				assert.Error(t, lvl.UnmarshalText([]byte(tt.text)), "Expected unmarshaling %q to fail.", tt.text)
			} else {
				assert.NoError(t, lvl.UnmarshalText([]byte(tt.text)), "Expected unmarshaling %q to succeed.", tt.text)
			}
			assert.Equal(t, tt.expect, lvl.Level(), "Unexpected level after unmarshaling.")
			lvl.SetLevel(InfoLevel)
		}

		// Test marshalling
		if tt.text != "" && !tt.err {
			lvl.SetLevel(tt.expect)
			marshaled, err := lvl.MarshalText()
			assert.NoError(t, err, `Unexpected error marshalling level "%v" to text.`, tt.expect)
			assert.Equal(t, tt.text, string(marshaled), "Expected marshaled text to match")
			assert.Equal(t, tt.text, lvl.String(), "Expected Stringer call to match")
		}
	}
}

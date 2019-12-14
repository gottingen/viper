

package vipercore_test

import (
	"testing"

	. "github.com/gottingen/viper/vipercore"
	"github.com/gottingen/viper/vipertest/observer"

	"github.com/stretchr/testify/assert"
)

func TestHooks(t *testing.T) {
	tests := []struct {
		entryLevel Level
		coreLevel  Level
		expectCall bool
	}{
		{DebugLevel, InfoLevel, false},
		{InfoLevel, InfoLevel, true},
		{WarnLevel, InfoLevel, true},
	}

	for _, tt := range tests {
		fac, logs := observer.New(tt.coreLevel)
		intField := makeInt64Field("foo", 42)
		ent := Entry{Message: "bar", Level: tt.entryLevel}

		var called int
		f := func(e Entry) error {
			called++
			assert.Equal(t, ent, e, "Hook called with unexpected Entry.")
			return nil
		}

		h := RegisterHooks(fac, f)
		if ce := h.With([]Field{intField}).Check(ent, nil); ce != nil {
			ce.Write()
		}

		if tt.expectCall {
			assert.Equal(t, 1, called, "Expected to call hook once.")
			assert.Equal(
				t,
				[]observer.LoggedEntry{{Entry: ent, Context: []Field{intField}}},
				logs.AllUntimed(),
				"Unexpected logs written out.",
			)
		} else {
			assert.Equal(t, 0, called, "Didn't expect to call hook.")
			assert.Equal(t, 0, logs.Len(), "Unexpected logs written out.")
		}
	}
}

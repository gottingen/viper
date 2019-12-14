

package vipercore

import (
	"sync"
	"testing"

	"github.com/gottingen/viper/internal/exit"

	"github.com/stretchr/testify/assert"
)

func TestPutNilEntry(t *testing.T) {
	// Pooling nil entries defeats the purpose.
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			putCheckedEntry(nil)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			ce := getCheckedEntry()
			assert.NotNil(t, ce, "Expected only non-nil CheckedEntries in pool.")
			assert.False(t, ce.dirty, "Unexpected dirty bit set.")
			assert.Nil(t, ce.ErrorOutput, "Non-nil ErrorOutput.")
			assert.Equal(t, WriteThenNoop, ce.should, "Unexpected terminal behavior.")
			assert.Equal(t, 0, len(ce.cores), "Expected empty slice of cores.")
			assert.True(t, cap(ce.cores) > 0, "Expected pooled CheckedEntries to pre-allocate slice of Cores.")
		}
	}()

	wg.Wait()
}

func TestEntryCaller(t *testing.T) {
	tests := []struct {
		caller EntryCaller
		full   string
		short  string
	}{
		{
			caller: NewEntryCaller(100, "/path/to/foo.go", 42, false),
			full:   "undefined",
			short:  "undefined",
		},
		{
			caller: NewEntryCaller(100, "/path/to/foo.go", 42, true),
			full:   "/path/to/foo.go:42",
			short:  "to/foo.go:42",
		},
		{
			caller: NewEntryCaller(100, "to/foo.go", 42, true),
			full:   "to/foo.go:42",
			short:  "to/foo.go:42",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.full, tt.caller.String(), "Unexpected string from EntryCaller.")
		assert.Equal(t, tt.full, tt.caller.FullPath(), "Unexpected FullPath from EntryCaller.")
		assert.Equal(t, tt.short, tt.caller.TrimmedPath(), "Unexpected TrimmedPath from EntryCaller.")
	}
}

func TestCheckedEntryWrite(t *testing.T) {
	// Nil checked entries are safe.
	var ce *CheckedEntry
	assert.NotPanics(t, func() { ce.Write() }, "Unexpected panic writing nil CheckedEntry.")

	// WriteThenPanic
	ce = ce.Should(Entry{}, WriteThenPanic)
	assert.Panics(t, func() { ce.Write() }, "Expected to panic when WriteThenPanic is set.")
	ce.reset()

	// WriteThenFatal
	ce = ce.Should(Entry{}, WriteThenFatal)
	stub := exit.WithStub(func() {
		ce.Write()
	})
	assert.True(t, stub.Exited, "Expected to exit when WriteThenFatal is set.")
	ce.reset()
}

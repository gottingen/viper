

package vipercore

import (
	"bytes"
	"errors"
	"testing"

	"io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/gottingen/viper/internal/vtest"
)

type writeSyncSpy struct {
	io.Writer
	vtest.Syncer
}

func requireWriteWorks(t testing.TB, ws WriteSyncer) {
	n, err := ws.Write([]byte("foo"))
	require.NoError(t, err, "Unexpected error writing to WriteSyncer.")
	require.Equal(t, 3, n, "Wrote an unexpected number of bytes.")
}

func TestAddSyncWriteSyncer(t *testing.T) {
	buf := &bytes.Buffer{}
	concrete := &writeSyncSpy{Writer: buf}
	ws := AddSync(concrete)
	requireWriteWorks(t, ws)

	require.NoError(t, ws.Sync(), "Unexpected error syncing a WriteSyncer.")
	require.True(t, concrete.Called(), "Expected to dispatch to concrete type's Sync method.")

	concrete.SetError(errors.New("fail"))
	assert.Error(t, ws.Sync(), "Expected to propagate errors from concrete type's Sync method.")
}

func TestAddSyncWriter(t *testing.T) {
	// If we pass a plain io.Writer, make sure that we still get a WriteSyncer
	// with a no-op Sync.
	buf := &bytes.Buffer{}
	ws := AddSync(buf)
	requireWriteWorks(t, ws)
	assert.NoError(t, ws.Sync(), "Unexpected error calling a no-op Sync method.")
}

func TestNewMultiWriteSyncerWorksForSingleWriter(t *testing.T) {
	w := &vtest.Buffer{}

	ws := NewMultiWriteSyncer(w)
	assert.Equal(t, w, ws, "Expected NewMultiWriteSyncer to return the same WriteSyncer object for a single argument.")

	ws.Sync()
	assert.True(t, w.Called(), "Expected Sync to be called on the created WriteSyncer")
}

func TestMultiWriteSyncerWritesBoth(t *testing.T) {
	first := &bytes.Buffer{}
	second := &bytes.Buffer{}
	ws := NewMultiWriteSyncer(AddSync(first), AddSync(second))

	msg := []byte("dumbledore")
	n, err := ws.Write(msg)
	require.NoError(t, err, "Expected successful buffer write")
	assert.Equal(t, len(msg), n)

	assert.Equal(t, msg, first.Bytes())
	assert.Equal(t, msg, second.Bytes())
}

func TestMultiWriteSyncerFailsWrite(t *testing.T) {
	ws := NewMultiWriteSyncer(AddSync(&vtest.FailWriter{}))
	_, err := ws.Write([]byte("test"))
	assert.Error(t, err, "Write error should propagate")
}

func TestMultiWriteSyncerFailsShortWrite(t *testing.T) {
	ws := NewMultiWriteSyncer(AddSync(&vtest.ShortWriter{}))
	n, err := ws.Write([]byte("test"))
	assert.NoError(t, err, "Expected fake-success from short write")
	assert.Equal(t, 3, n, "Expected byte count to return from underlying writer")
}

func TestWritestoAllSyncs_EvenIfFirstErrors(t *testing.T) {
	failer := &vtest.FailWriter{}
	second := &bytes.Buffer{}
	ws := NewMultiWriteSyncer(AddSync(failer), AddSync(second))

	_, err := ws.Write([]byte("fail"))
	assert.Error(t, err, "Expected error from call to a writer that failed")
	assert.Equal(t, []byte("fail"), second.Bytes(), "Expected second sink to be written after first error")
}

func TestMultiWriteSyncerSync_PropagatesErrors(t *testing.T) {
	badsink := &vtest.Buffer{}
	badsink.SetError(errors.New("sink is full"))
	ws := NewMultiWriteSyncer(&vtest.Discarder{}, badsink)

	assert.Error(t, ws.Sync(), "Expected sync error to propagate")
}

func TestMultiWriteSyncerSync_NoErrorsOnDiscard(t *testing.T) {
	ws := NewMultiWriteSyncer(&vtest.Discarder{})
	assert.NoError(t, ws.Sync(), "Expected error-free sync to /dev/null")
}

func TestMultiWriteSyncerSync_AllCalled(t *testing.T) {
	failed, second := &vtest.Buffer{}, &vtest.Buffer{}

	failed.SetError(errors.New("disposal broken"))
	ws := NewMultiWriteSyncer(failed, second)

	assert.Error(t, ws.Sync(), "Expected first sink to fail")
	assert.True(t, failed.Called(), "Expected first sink to have Sync method called.")
	assert.True(t, second.Called(), "Expected call to Sync even with first failure.")
}



package vipertest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncer(t *testing.T) {
	err := errors.New("sentinel")
	s := &Syncer{}
	s.SetError(err)
	assert.Equal(t, err, s.Sync(), "Expected Sync to fail with provided error.")
	assert.True(t, s.Called(), "Expected to record that Sync was called.")
}

func TestDiscarder(t *testing.T) {
	d := &Discarder{}
	payload := []byte("foo")
	n, err := d.Write(payload)
	assert.NoError(t, err, "Unexpected error writing to Discarder.")
	assert.Equal(t, len(payload), n, "Wrong number of bytes written.")
}

func TestFailWriter(t *testing.T) {
	w := &FailWriter{}
	payload := []byte("foo")
	n, err := w.Write(payload)
	assert.Error(t, err, "Expected an error writing to FailWriter.")
	assert.Equal(t, len(payload), n, "Wrong number of bytes written.")
}

func TestShortWriter(t *testing.T) {
	w := &ShortWriter{}
	payload := []byte("foo")
	n, err := w.Write(payload)
	assert.NoError(t, err, "Unexpected error writing to ShortWriter.")
	assert.Equal(t, len(payload)-1, n, "Wrong number of bytes written.")
}

func TestBuffer(t *testing.T) {
	buf := &Buffer{}
	buf.WriteString("foo\n")
	buf.WriteString("bar\n")
	assert.Equal(t, []string{"foo", "bar"}, buf.Lines(), "Unexpected output from Lines.")
	assert.Equal(t, "foo\nbar", buf.Stripped(), "Unexpected output from Stripped.")
}

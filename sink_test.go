

package viper

import (
	"bytes"
	"io/ioutil"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gottingen/viper/vipercore"
)

func TestRegisterSink(t *testing.T) {
	const (
		memScheme = "m"
		nopScheme = "no-op.1234"
	)
	var memCalls, nopCalls int

	buf := bytes.NewBuffer(nil)
	memFactory := func(u *url.URL) (Sink, error) {
		assert.Equal(t, u.Scheme, memScheme, "Scheme didn't match registration.")
		memCalls++
		return nopCloserSink{vipercore.AddSync(buf)}, nil
	}
	nopFactory := func(u *url.URL) (Sink, error) {
		assert.Equal(t, u.Scheme, nopScheme, "Scheme didn't match registration.")
		nopCalls++
		return nopCloserSink{vipercore.AddSync(ioutil.Discard)}, nil
	}

	defer resetSinkRegistry()

	require.NoError(t, RegisterSink(strings.ToUpper(memScheme), memFactory), "Failed to register scheme %q.", memScheme)
	require.NoError(t, RegisterSink(nopScheme, nopFactory), "Failed to register scheme %q.", memScheme)

	sink, close, err := Open(
		memScheme+"://somewhere",
		nopScheme+"://somewhere-else",
	)
	assert.NoError(t, err, "Unexpected error opening URLs with registered schemes.")

	defer close()

	assert.Equal(t, 1, memCalls, "Unexpected number of calls to memory factory.")
	assert.Equal(t, 1, nopCalls, "Unexpected number of calls to no-op factory.")

	_, err = sink.Write([]byte("foo"))
	assert.NoError(t, err, "Failed to write to combined WriteSyncer.")
	assert.Equal(t, "foo", buf.String(), "Unexpected buffer contents.")
}

func TestRegisterSinkErrors(t *testing.T) {
	nopFactory := func(_ *url.URL) (Sink, error) {
		return nopCloserSink{vipercore.AddSync(ioutil.Discard)}, nil
	}
	tests := []struct {
		scheme string
		err    string
	}{
		{"", "empty string"},
		{"FILE", "already registered"},
		{"42", "not a valid scheme"},
		{"http*", "not a valid scheme"},
	}

	for _, tt := range tests {
		t.Run("scheme-"+tt.scheme, func(t *testing.T) {
			defer resetSinkRegistry()

			err := RegisterSink(tt.scheme, nopFactory)
			if assert.Error(t, err, "expected error") {
				assert.Contains(t, err.Error(), tt.err, "unexpected error")
			}
		})
	}
}



package viper_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/gottingen/viper"
	"github.com/gottingen/viper/vipercore"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHandler() (AtomicLevel, *Logger) {
	lvl := NewAtomicLevel()
	logger := New(vipercore.NewNopCore())
	return lvl, logger
}

func assertCodeOK(t testing.TB, code int) {
	assert.Equal(t, http.StatusOK, code, "Unexpected response status code.")
}

func assertCodeBadRequest(t testing.TB, code int) {
	assert.Equal(t, http.StatusBadRequest, code, "Unexpected response status code.")
}

func assertCodeMethodNotAllowed(t testing.TB, code int) {
	assert.Equal(t, http.StatusMethodNotAllowed, code, "Unexpected response status code.")
}

func assertResponse(t testing.TB, expectedLevel vipercore.Level, actualBody string) {
	assert.Equal(t, fmt.Sprintf(`{"level":"%s"}`, expectedLevel)+"\n", actualBody, "Unexpected response body.")
}

func assertJSONError(t testing.TB, body string) {
	// Don't need to test exact error message, but one should be present.
	var payload map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(body), &payload), "Expected error response to be JSON.")

	msg, ok := payload["error"]
	require.True(t, ok, "Error message is an unexpected type.")
	assert.NotEqual(t, "", msg, "Expected an error message in response.")
}

func makeRequest(t testing.TB, method string, handler http.Handler, reader io.Reader) (int, string) {
	ts := httptest.NewServer(handler)
	defer ts.Close()

	req, err := http.NewRequest(method, ts.URL, reader)
	require.NoError(t, err, "Error constructing %s request.", method)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err, "Error making %s request.", method)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(t, err, "Error reading request body.")

	return res.StatusCode, string(body)
}

func TestHTTPHandlerGetLevel(t *testing.T) {
	lvl, _ := newHandler()
	code, body := makeRequest(t, "GET", lvl, nil)
	assertCodeOK(t, code)
	assertResponse(t, lvl.Level(), body)
}

func TestHTTPHandlerPutLevel(t *testing.T) {
	lvl, _ := newHandler()

	code, body := makeRequest(t, "PUT", lvl, strings.NewReader(`{"level":"warn"}`))

	assertCodeOK(t, code)
	assertResponse(t, lvl.Level(), body)
}

func TestHTTPHandlerPutUnrecognizedLevel(t *testing.T) {
	lvl, _ := newHandler()
	code, body := makeRequest(t, "PUT", lvl, strings.NewReader(`{"level":"unrecognized-level"}`))
	assertCodeBadRequest(t, code)
	assertJSONError(t, body)
}

func TestHTTPHandlerNotJSON(t *testing.T) {
	lvl, _ := newHandler()
	code, body := makeRequest(t, "PUT", lvl, strings.NewReader(`{`))
	assertCodeBadRequest(t, code)
	assertJSONError(t, body)
}

func TestHTTPHandlerNoLevelSpecified(t *testing.T) {
	lvl, _ := newHandler()
	code, body := makeRequest(t, "PUT", lvl, strings.NewReader(`{}`))
	assertCodeBadRequest(t, code)
	assertJSONError(t, body)
}

func TestHTTPHandlerMethodNotAllowed(t *testing.T) {
	lvl, _ := newHandler()
	code, body := makeRequest(t, "POST", lvl, strings.NewReader(`{`))
	assertCodeMethodNotAllowed(t, code)
	assertJSONError(t, body)
}

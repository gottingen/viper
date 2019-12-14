

package vipercore_test

import (
	"errors"
	"testing"

	"github.com/gottingen/viper/internal/vtest"
	. "github.com/gottingen/viper/vipercore"
	"github.com/gottingen/viper/vipertest/observer"

	"github.com/stretchr/testify/assert"
)

func withTee(f func(core Core, debugLogs, warnLogs *observer.ObservedLogs)) {
	debugLogger, debugLogs := observer.New(DebugLevel)
	warnLogger, warnLogs := observer.New(WarnLevel)
	tee := NewTee(debugLogger, warnLogger)
	f(tee, debugLogs, warnLogs)
}

func TestTeeUnusualInput(t *testing.T) {
	// Verify that Tee handles receiving one and no inputs correctly.
	t.Run("one input", func(t *testing.T) {
		obs, _ := observer.New(DebugLevel)
		assert.Equal(t, obs, NewTee(obs), "Expected to return single inputs unchanged.")
	})
	t.Run("no input", func(t *testing.T) {
		assert.Equal(t, NewNopCore(), NewTee(), "Expected to return NopCore.")
	})
}

func TestTeeCheck(t *testing.T) {
	withTee(func(tee Core, debugLogs, warnLogs *observer.ObservedLogs) {
		debugEntry := Entry{Level: DebugLevel, Message: "log-at-debug"}
		infoEntry := Entry{Level: InfoLevel, Message: "log-at-info"}
		warnEntry := Entry{Level: WarnLevel, Message: "log-at-warn"}
		errorEntry := Entry{Level: ErrorLevel, Message: "log-at-error"}
		for _, ent := range []Entry{debugEntry, infoEntry, warnEntry, errorEntry} {
			if ce := tee.Check(ent, nil); ce != nil {
				ce.Write()
			}
		}

		assert.Equal(t, []observer.LoggedEntry{
			{Entry: debugEntry, Context: []Field{}},
			{Entry: infoEntry, Context: []Field{}},
			{Entry: warnEntry, Context: []Field{}},
			{Entry: errorEntry, Context: []Field{}},
		}, debugLogs.All())

		assert.Equal(t, []observer.LoggedEntry{
			{Entry: warnEntry, Context: []Field{}},
			{Entry: errorEntry, Context: []Field{}},
		}, warnLogs.All())
	})
}

func TestTeeWrite(t *testing.T) {
	// Calling the tee's Write method directly should always log, regardless of
	// the configured level.
	withTee(func(tee Core, debugLogs, warnLogs *observer.ObservedLogs) {
		debugEntry := Entry{Level: DebugLevel, Message: "log-at-debug"}
		warnEntry := Entry{Level: WarnLevel, Message: "log-at-warn"}
		for _, ent := range []Entry{debugEntry, warnEntry} {
			tee.Write(ent, nil)
		}

		for _, logs := range []*observer.ObservedLogs{debugLogs, warnLogs} {
			assert.Equal(t, []observer.LoggedEntry{
				{Entry: debugEntry, Context: []Field{}},
				{Entry: warnEntry, Context: []Field{}},
			}, logs.All())
		}
	})
}

func TestTeeWith(t *testing.T) {
	withTee(func(tee Core, debugLogs, warnLogs *observer.ObservedLogs) {
		f := makeInt64Field("k", 42)
		tee = tee.With([]Field{f})
		ent := Entry{Level: WarnLevel, Message: "log-at-warn"}
		if ce := tee.Check(ent, nil); ce != nil {
			ce.Write()
		}

		for _, logs := range []*observer.ObservedLogs{debugLogs, warnLogs} {
			assert.Equal(t, []observer.LoggedEntry{
				{Entry: ent, Context: []Field{f}},
			}, logs.All())
		}
	})
}

func TestTeeEnabled(t *testing.T) {
	infoLogger, _ := observer.New(InfoLevel)
	warnLogger, _ := observer.New(WarnLevel)
	tee := NewTee(infoLogger, warnLogger)
	tests := []struct {
		lvl     Level
		enabled bool
	}{
		{DebugLevel, false},
		{InfoLevel, true},
		{WarnLevel, true},
		{ErrorLevel, true},
		{DPanicLevel, true},
		{PanicLevel, true},
		{FatalLevel, true},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.enabled, tee.Enabled(tt.lvl), "Unexpected Enabled result for level %s.", tt.lvl)
	}
}

func TestTeeSync(t *testing.T) {
	infoLogger, _ := observer.New(InfoLevel)
	warnLogger, _ := observer.New(WarnLevel)
	tee := NewTee(infoLogger, warnLogger)
	assert.NoError(t, tee.Sync(), "Unexpected error from Syncing a tee.")

	sink := &vtest.Discarder{}
	err := errors.New("failed")
	sink.SetError(err)

	noSync := NewCore(
		NewJSONEncoder(testEncoderConfig()),
		sink,
		DebugLevel,
	)
	tee = NewTee(tee, noSync)
	assert.Equal(t, err, tee.Sync(), "Expected an error when part of tee can't Sync.")
}

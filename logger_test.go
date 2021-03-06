

package viper

import (
	"errors"
	"sync"
	"testing"

	"github.com/gottingen/viper/internal/exit"
	"github.com/gottingen/viper/internal/vtest"
	"github.com/gottingen/viper/vipercore"
	"github.com/gottingen/viper/vipertest/observer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/gottingen/atomic"
)

func makeCountingHook() (func(vipercore.Entry) error, *atomic.Int64) {
	count := &atomic.Int64{}
	h := func(vipercore.Entry) error {
		count.Inc()
		return nil
	}
	return h, count
}

func TestLoggerAtomicLevel(t *testing.T) {
	// Test that the dynamic level applies to all ancestors and descendants.
	dl := NewAtomicLevel()

	withLogger(t, dl, nil, func(grandparent *Logger, _ *observer.ObservedLogs) {
		parent := grandparent.With(Int("generation", 1))
		child := parent.With(Int("generation", 2))

		tests := []struct {
			setLevel  vipercore.Level
			testLevel vipercore.Level
			enabled   bool
		}{
			{DebugLevel, DebugLevel, true},
			{InfoLevel, DebugLevel, false},
			{WarnLevel, PanicLevel, true},
		}

		for _, tt := range tests {
			dl.SetLevel(tt.setLevel)
			for _, logger := range []*Logger{grandparent, parent, child} {
				if tt.enabled {
					assert.NotNil(
						t,
						logger.Check(tt.testLevel, ""),
						"Expected level %s to be enabled after setting level %s.", tt.testLevel, tt.setLevel,
					)
				} else {
					assert.Nil(
						t,
						logger.Check(tt.testLevel, ""),
						"Expected level %s to be enabled after setting level %s.", tt.testLevel, tt.setLevel,
					)
				}
			}
		}
	})
}

func TestLoggerInitialFields(t *testing.T) {
	fieldOpts := opts(Fields(Int("foo", 42), String("bar", "baz")))
	withLogger(t, DebugLevel, fieldOpts, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Info("")
		assert.Equal(
			t,
			observer.LoggedEntry{Context: []Field{Int("foo", 42), String("bar", "baz")}},
			logs.AllUntimed()[0],
			"Unexpected output with initial fields set.",
		)
	})
}

func TestLoggerWith(t *testing.T) {
	fieldOpts := opts(Fields(Int("foo", 42)))
	withLogger(t, DebugLevel, fieldOpts, func(logger *Logger, logs *observer.ObservedLogs) {
		// Child loggers should have copy-on-write semantics, so two children
		// shouldn't stomp on each other's fields or affect the parent's fields.
		logger.With(String("one", "two")).Info("")
		logger.With(String("three", "four")).Info("")
		logger.Info("")

		assert.Equal(t, []observer.LoggedEntry{
			{Context: []Field{Int("foo", 42), String("one", "two")}},
			{Context: []Field{Int("foo", 42), String("three", "four")}},
			{Context: []Field{Int("foo", 42)}},
		}, logs.AllUntimed(), "Unexpected cross-talk between child loggers.")
	})
}

func TestLoggerLogPanic(t *testing.T) {
	for _, tt := range []struct {
		do       func(*Logger)
		should   bool
		expected string
	}{
		{func(logger *Logger) { logger.Check(PanicLevel, "bar").Write() }, true, "bar"},
		{func(logger *Logger) { logger.Panic("baz") }, true, "baz"},
	} {
		withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
			if tt.should {
				assert.Panics(t, func() { tt.do(logger) }, "Expected panic")
			} else {
				assert.NotPanics(t, func() { tt.do(logger) }, "Expected no panic")
			}

			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[0].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				vipercore.Entry{Message: tt.expected, Level: PanicLevel},
				output[0].Entry,
				"Unexpected output from panic-level Log.",
			)
		})
	}
}

func TestLoggerLogFatal(t *testing.T) {
	for _, tt := range []struct {
		do       func(*Logger)
		expected string
	}{
		{func(logger *Logger) { logger.Check(FatalLevel, "bar").Write() }, "bar"},
		{func(logger *Logger) { logger.Fatal("baz") }, "baz"},
	} {
		withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
			stub := exit.WithStub(func() {
				tt.do(logger)
			})
			assert.True(t, stub.Exited, "Expected Fatal logger call to terminate process.")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[0].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				vipercore.Entry{Message: tt.expected, Level: FatalLevel},
				output[0].Entry,
				"Unexpected output from fatal-level Log.",
			)
		})
	}
}

func TestLoggerLeveledMethods(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		tests := []struct {
			method        func(string, ...Field)
			expectedLevel vipercore.Level
		}{
			{logger.Debug, DebugLevel},
			{logger.Info, InfoLevel},
			{logger.Warn, WarnLevel},
			{logger.Error, ErrorLevel},
			{logger.DPanic, DPanicLevel},
		}
		for i, tt := range tests {
			tt.method("")
			output := logs.AllUntimed()
			assert.Equal(t, i+1, len(output), "Unexpected number of logs.")
			assert.Equal(t, 0, len(output[i].Context), "Unexpected context on first log.")
			assert.Equal(
				t,
				vipercore.Entry{Level: tt.expectedLevel},
				output[i].Entry,
				"Unexpected output from %s-level logger method.", tt.expectedLevel)
		}
	})
}

func TestLoggerAlwaysPanics(t *testing.T) {
	// Users can disable writing out panic-level logs, but calls to logger.Panic()
	// should still call panic().
	withLogger(t, FatalLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		msg := "Even if output is disabled, logger.Panic should always panic."
		assert.Panics(t, func() { logger.Panic("foo") }, msg)
		assert.Panics(t, func() {
			if ce := logger.Check(PanicLevel, "foo"); ce != nil {
				ce.Write()
			}
		}, msg)
		assert.Equal(t, 0, logs.Len(), "Panics shouldn't be written out if PanicLevel is disabled.")
	})
}

func TestLoggerAlwaysFatals(t *testing.T) {
	// Users can disable writing out fatal-level logs, but calls to logger.Fatal()
	// should still terminate the process.
	withLogger(t, FatalLevel+1, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		stub := exit.WithStub(func() { logger.Fatal("") })
		assert.True(t, stub.Exited, "Expected calls to logger.Fatal to terminate process.")

		stub = exit.WithStub(func() {
			if ce := logger.Check(FatalLevel, ""); ce != nil {
				ce.Write()
			}
		})
		assert.True(t, stub.Exited, "Expected calls to logger.Check(FatalLevel, ...) to terminate process.")

		assert.Equal(t, 0, logs.Len(), "Shouldn't write out logs when fatal-level logging is disabled.")
	})
}

func TestLoggerDPanic(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		assert.NotPanics(t, func() { logger.DPanic("") })
		assert.Equal(
			t,
			[]observer.LoggedEntry{{Entry: vipercore.Entry{Level: DPanicLevel}, Context: []Field{}}},
			logs.AllUntimed(),
			"Unexpected log output from DPanic in production mode.",
		)
	})
	withLogger(t, DebugLevel, opts(Development()), func(logger *Logger, logs *observer.ObservedLogs) {
		assert.Panics(t, func() { logger.DPanic("") })
		assert.Equal(
			t,
			[]observer.LoggedEntry{{Entry: vipercore.Entry{Level: DPanicLevel}, Context: []Field{}}},
			logs.AllUntimed(),
			"Unexpected log output from DPanic in development mode.",
		)
	})
}

func TestLoggerNoOpsDisabledLevels(t *testing.T) {
	withLogger(t, WarnLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Info("silence!")
		assert.Equal(
			t,
			[]observer.LoggedEntry{},
			logs.AllUntimed(),
			"Expected logging at a disabled level to produce no output.",
		)
	})
}

func TestLoggerNames(t *testing.T) {
	tests := []struct {
		names    []string
		expected string
	}{
		{nil, ""},
		{[]string{""}, ""},
		{[]string{"foo"}, "foo"},
		{[]string{"foo", ""}, "foo"},
		{[]string{"foo", "bar"}, "foo.bar"},
		{[]string{"foo.bar", "baz"}, "foo.bar.baz"},
		// Garbage in, garbage out.
		{[]string{"foo.", "bar"}, "foo..bar"},
		{[]string{"foo", ".bar"}, "foo..bar"},
		{[]string{"foo.", ".bar"}, "foo...bar"},
	}

	for _, tt := range tests {
		withLogger(t, DebugLevel, nil, func(log *Logger, logs *observer.ObservedLogs) {
			for _, n := range tt.names {
				log = log.Named(n)
			}
			log.Info("")
			require.Equal(t, 1, logs.Len(), "Expected only one log entry to be written.")
			assert.Equal(t, tt.expected, logs.AllUntimed()[0].Entry.LoggerName, "Unexpected logger name.")
		})
		withSugar(t, DebugLevel, nil, func(log *SugaredLogger, logs *observer.ObservedLogs) {
			for _, n := range tt.names {
				log = log.Named(n)
			}
			log.Infow("")
			require.Equal(t, 1, logs.Len(), "Expected only one log entry to be written.")
			assert.Equal(t, tt.expected, logs.AllUntimed()[0].Entry.LoggerName, "Unexpected logger name.")
		})
	}
}

func TestLoggerWriteFailure(t *testing.T) {
	errSink := &vtest.Buffer{}
	logger := New(
		vipercore.NewCore(
			vipercore.NewJSONEncoder(NewProductionConfig().EncoderConfig),
			vipercore.Lock(vipercore.AddSync(vtest.FailWriter{})),
			DebugLevel,
		),
		ErrorOutput(errSink),
	)

	logger.Info("foo")
	// Should log the error.
	assert.Regexp(t, `write error: failed`, errSink.Stripped(), "Expected to log the error to the error output.")
	assert.True(t, errSink.Called(), "Expected logging an internal error to call Sync the error sink.")
}

func TestLoggerSync(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, _ *observer.ObservedLogs) {
		assert.NoError(t, logger.Sync(), "Expected syncing a test logger to succeed.")
		assert.NoError(t, logger.Sugar().Sync(), "Expected syncing a sugared logger to succeed.")
	})
}

func TestLoggerSyncFail(t *testing.T) {
	noSync := &vtest.Buffer{}
	err := errors.New("fail")
	noSync.SetError(err)
	logger := New(vipercore.NewCore(
		vipercore.NewJSONEncoder(vipercore.EncoderConfig{}),
		noSync,
		DebugLevel,
	))
	assert.Equal(t, err, logger.Sync(), "Expected Logger.Sync to propagate errors.")
	assert.Equal(t, err, logger.Sugar().Sync(), "Expected SugaredLogger.Sync to propagate errors.")
}

func TestLoggerAddCaller(t *testing.T) {
	tests := []struct {
		options []Option
		pat     string
	}{
		{opts(AddCaller()), `.+/logger_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(-1)), `.+/viper/logger_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1)), `.+/viper/common_test.go:[\d]+$`},
		{opts(AddCaller(), AddCallerSkip(1), AddCallerSkip(3)), `.+/src/runtime/.*:[\d]+$`},
	}
	for _, tt := range tests {
		withLogger(t, DebugLevel, tt.options, func(logger *Logger, logs *observer.ObservedLogs) {
			// Make sure that sugaring and desugaring resets caller skip properly.
			logger = logger.Sugar().Desugar()
			logger.Info("")
			output := logs.AllUntimed()
			assert.Equal(t, 1, len(output), "Unexpected number of logs written out.")
			assert.Regexp(
				t,
				tt.pat,
				output[0].Entry.Caller,
				"Expected to find package name and file name in output.",
			)
		})
	}
}

func TestLoggerAddCallerFail(t *testing.T) {
	errBuf := &vtest.Buffer{}
	withLogger(t, DebugLevel, opts(AddCaller(), ErrorOutput(errBuf)), func(log *Logger, logs *observer.ObservedLogs) {
		log.callerSkip = 1e3
		log.Info("Failure.")
		assert.Regexp(
			t,
			`Logger.check error: failed to get caller`,
			errBuf.String(),
			"Didn't find expected failure message.",
		)
		assert.Equal(
			t,
			logs.AllUntimed()[0].Entry.Message,
			"Failure.",
			"Expected original message to survive failures in runtime.Caller.")
	})
}

func TestLoggerReplaceCore(t *testing.T) {
	replace := WrapCore(func(vipercore.Core) vipercore.Core {
		return vipercore.NewNopCore()
	})
	withLogger(t, DebugLevel, opts(replace), func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Debug("")
		logger.Info("")
		logger.Warn("")
		assert.Equal(t, 0, logs.Len(), "Expected no-op core to write no logs.")
	})
}

func TestLoggerHooks(t *testing.T) {
	hook, seen := makeCountingHook()
	withLogger(t, DebugLevel, opts(Hooks(hook)), func(logger *Logger, logs *observer.ObservedLogs) {
		logger.Debug("")
		logger.Info("")
	})
	assert.Equal(t, int64(2), seen.Load(), "Hook saw an unexpected number of logs.")
}

func TestLoggerConcurrent(t *testing.T) {
	withLogger(t, DebugLevel, nil, func(logger *Logger, logs *observer.ObservedLogs) {
		child := logger.With(String("foo", "bar"))

		wg := &sync.WaitGroup{}
		runConcurrently(5, 10, wg, func() {
			logger.Info("", String("foo", "bar"))
		})
		runConcurrently(5, 10, wg, func() {
			child.Info("")
		})

		wg.Wait()

		// Make sure the output doesn't contain interspersed entries.
		assert.Equal(t, 100, logs.Len(), "Unexpected number of logs written out.")
		for _, obs := range logs.AllUntimed() {
			assert.Equal(
				t,
				observer.LoggedEntry{
					Entry:   vipercore.Entry{Level: InfoLevel},
					Context: []Field{String("foo", "bar")},
				},
				obs,
				"Unexpected log output.",
			)
		}
	})
}

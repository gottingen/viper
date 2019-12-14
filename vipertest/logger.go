// Copyright (c) 2017 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package vipertest

import (
	"bytes"

	"github.com/gottingen/viper"
	"github.com/gottingen/viper/vipercore"
)

// LoggerOption configures the test logger built by NewLogger.
type LoggerOption interface {
	applyLoggerOption(*loggerOptions)
}

type loggerOptions struct {
	Level      vipercore.LevelEnabler
	viperOptions []viper.Option
}

type loggerOptionFunc func(*loggerOptions)

func (f loggerOptionFunc) applyLoggerOption(opts *loggerOptions) {
	f(opts)
}

// Level controls which messages are logged by a test Logger built by
// NewLogger.
func Level(enab vipercore.LevelEnabler) LoggerOption {
	return loggerOptionFunc(func(opts *loggerOptions) {
		opts.Level = enab
	})
}

// WrapOptions adds viper.Option's to a test Logger built by NewLogger.
func WrapOptions(viperOpts ...viper.Option) LoggerOption {
	return loggerOptionFunc(func(opts *loggerOptions) {
		opts.viperOptions = viperOpts
	})
}

// NewLogger builds a new Logger that logs all messages to the given
// testing.TB.
//
//   logger := vipertest.NewLogger(t)
//
// Use this with a *testing.T or *testing.B to get logs which get printed only
// if a test fails or if you ran go test -v.
//
// The returned logger defaults to logging debug level messages and above.
// This may be changed by passing a vipertest.Level during construction.
//
//   logger := vipertest.NewLogger(t, vipertest.Level(viper.WarnLevel))
//
// You may also pass viper.Option's to customize test logger.
//
//   logger := vipertest.NewLogger(t, vipertest.WrapOptions(viper.AddCaller()))
func NewLogger(t TestingT, opts ...LoggerOption) *viper.Logger {
	cfg := loggerOptions{
		Level: vipercore.DebugLevel,
	}
	for _, o := range opts {
		o.applyLoggerOption(&cfg)
	}

	writer := newTestingWriter(t)
	viperOptions := []viper.Option{
		// Send viper errors to the same writer and mark the test as failed if
		// that happens.
		viper.ErrorOutput(writer.WithMarkFailed(true)),
	}
	viperOptions = append(viperOptions, cfg.viperOptions...)

	return viper.New(
		vipercore.NewCore(
			vipercore.NewConsoleEncoder(viper.NewDevelopmentEncoderConfig()),
			writer,
			cfg.Level,
		),
		viperOptions...,
	)
}

// testingWriter is a WriteSyncer that writes to the given testing.TB.
type testingWriter struct {
	t TestingT

	// If true, the test will be marked as failed if this testingWriter is
	// ever used.
	markFailed bool
}

func newTestingWriter(t TestingT) testingWriter {
	return testingWriter{t: t}
}

// WithMarkFailed returns a copy of this testingWriter with markFailed set to
// the provided value.
func (w testingWriter) WithMarkFailed(v bool) testingWriter {
	w.markFailed = v
	return w
}

func (w testingWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	// Strip trailing newline because t.Log always adds one.
	p = bytes.TrimRight(p, "\n")

	// Note: t.Log is safe for concurrent use.
	w.t.Logf("%s", p)
	if w.markFailed {
		w.t.Fail()
	}

	return n, nil
}

func (w testingWriter) Sync() error {
	return nil
}

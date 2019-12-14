

package viper

import (
	"github.com/gottingen/viper/vipercore"
	"github.com/gottingen/atomic"
)

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = vipercore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = vipercore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = vipercore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = vipercore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = vipercore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = vipercore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = vipercore.FatalLevel
)

// LevelEnablerFunc is a convenient way to implement vipercore.LevelEnabler with
// an anonymous function.
//
// It's particularly useful when splitting log output between different
// outputs (e.g., standard error and standard out). For sample code, see the
// package-level AdvancedConfiguration example.
type LevelEnablerFunc func(vipercore.Level) bool

// Enabled calls the wrapped function.
func (f LevelEnablerFunc) Enabled(lvl vipercore.Level) bool { return f(lvl) }

// An AtomicLevel is an atomically changeable, dynamic logging level. It lets
// you safely change the log level of a tree of loggers (the root logger and
// any children created by adding context) at runtime.
//
// The AtomicLevel itself is an http.Handler that serves a JSON endpoint to
// alter its level.
//
// AtomicLevels must be created with the NewAtomicLevel constructor to allocate
// their internal atomic pointer.
type AtomicLevel struct {
	l *atomic.Int32
}

// NewAtomicLevel creates an AtomicLevel with InfoLevel and above logging
// enabled.
func NewAtomicLevel() AtomicLevel {
	return AtomicLevel{
		l: atomic.NewInt32(int32(InfoLevel)),
	}
}

// NewAtomicLevelAt is a convenience function that creates an AtomicLevel
// and then calls SetLevel with the given level.
func NewAtomicLevelAt(l vipercore.Level) AtomicLevel {
	a := NewAtomicLevel()
	a.SetLevel(l)
	return a
}

// Enabled implements the vipercore.LevelEnabler interface, which allows the
// AtomicLevel to be used in place of traditional static levels.
func (lvl AtomicLevel) Enabled(l vipercore.Level) bool {
	return lvl.Level().Enabled(l)
}

// Level returns the minimum enabled log level.
func (lvl AtomicLevel) Level() vipercore.Level {
	return vipercore.Level(int8(lvl.l.Load()))
}

// SetLevel alters the logging level.
func (lvl AtomicLevel) SetLevel(l vipercore.Level) {
	lvl.l.Store(int32(l))
}

// String returns the string representation of the underlying Level.
func (lvl AtomicLevel) String() string {
	return lvl.Level().String()
}

// UnmarshalText unmarshals the text to an AtomicLevel. It uses the same text
// representations as the static vipercore.Levels ("debug", "info", "warn",
// "error", "dpanic", "panic", and "fatal").
func (lvl *AtomicLevel) UnmarshalText(text []byte) error {
	if lvl.l == nil {
		lvl.l = &atomic.Int32{}
	}

	var l vipercore.Level
	if err := l.UnmarshalText(text); err != nil {
		return err
	}

	lvl.SetLevel(l)
	return nil
}

// MarshalText marshals the AtomicLevel to a byte slice. It uses the same
// text representation as the static vipercore.Levels ("debug", "info", "warn",
// "error", "dpanic", "panic", and "fatal").
func (lvl AtomicLevel) MarshalText() (text []byte, err error) {
	return lvl.Level().MarshalText()
}

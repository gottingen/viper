

package viper

import "github.com/gottingen/viper/vipercore"

// An Option configures a Logger.
type Option interface {
	apply(*Logger)
}

// optionFunc wraps a func so it satisfies the Option interface.
type optionFunc func(*Logger)

func (f optionFunc) apply(log *Logger) {
	f(log)
}

// WrapCore wraps or replaces the Logger's underlying vipercore.Core.
func WrapCore(f func(vipercore.Core) vipercore.Core) Option {
	return optionFunc(func(log *Logger) {
		log.core = f(log.core)
	})
}

// Hooks registers functions which will be called each time the Logger writes
// out an Entry. Repeated use of Hooks is additive.
//
// Hooks are useful for simple side effects, like capturing metrics for the
// number of emitted logs. More complex side effects, including anything that
// requires access to the Entry's structured fields, should be implemented as
// a vipercore.Core instead. See vipercore.RegisterHooks for details.
func Hooks(hooks ...func(vipercore.Entry) error) Option {
	return optionFunc(func(log *Logger) {
		log.core = vipercore.RegisterHooks(log.core, hooks...)
	})
}

// Fields adds fields to the Logger.
func Fields(fs ...Field) Option {
	return optionFunc(func(log *Logger) {
		log.core = log.core.With(fs)
	})
}

// ErrorOutput sets the destination for errors generated by the Logger. Note
// that this option only affects internal errors; for sample code that sends
// error-level logs to a different location from info- and debug-level logs,
// see the package-level AdvancedConfiguration example.
//
// The supplied WriteSyncer must be safe for concurrent use. The Open and
// vipercore.Lock functions are the simplest ways to protect files with a mutex.
func ErrorOutput(w vipercore.WriteSyncer) Option {
	return optionFunc(func(log *Logger) {
		log.errorOutput = w
	})
}

// Development puts the logger in development mode, which makes DPanic-level
// logs panic instead of simply logging an error.
func Development() Option {
	return optionFunc(func(log *Logger) {
		log.development = true
	})
}

// AddCaller configures the Logger to annotate each message with the filename
// and line number of viper's caller.
func AddCaller() Option {
	return optionFunc(func(log *Logger) {
		log.addCaller = true
	})
}

// AddCallerSkip increases the number of callers skipped by caller annotation
// (as enabled by the AddCaller option). When building wrappers around the
// Logger and SugaredLogger, supplying this Option prevents viper from always
// reporting the wrapper code as the caller.
func AddCallerSkip(skip int) Option {
	return optionFunc(func(log *Logger) {
		log.callerSkip += skip
	})
}

// AddStacktrace configures the Logger to record a stack trace for all messages at
// or above a given level.
func AddStacktrace(lvl vipercore.LevelEnabler) Option {
	return optionFunc(func(log *Logger) {
		log.addStack = lvl
	})
}

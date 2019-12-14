

// package vipergrpc provides a logger that is compatible with grpclog.
package vipergrpc // import "github.com/gottingen/viper/vipergrpc"

import "github.com/gottingen/viper"

// An Option overrides a Logger's default configuration.
type Option interface {
	apply(*Logger)
}

type optionFunc func(*Logger)

func (f optionFunc) apply(log *Logger) {
	f(log)
}

// WithDebug configures a Logger to print at viper's DebugLevel instead of
// InfoLevel.
func WithDebug() Option {
	return optionFunc(func(logger *Logger) {
		logger.print = (*viper.SugaredLogger).Debug
		logger.printf = (*viper.SugaredLogger).Debugf
	})
}

// NewLogger returns a new Logger.
//
// By default, Loggers print at viper's InfoLevel.
func NewLogger(l *viper.Logger, options ...Option) *Logger {
	logger := &Logger{
		log:    l.Sugar(),
		fatal:  (*viper.SugaredLogger).Fatal,
		fatalf: (*viper.SugaredLogger).Fatalf,
		print:  (*viper.SugaredLogger).Info,
		printf: (*viper.SugaredLogger).Infof,
	}
	for _, option := range options {
		option.apply(logger)
	}
	return logger
}

// Logger adapts viper's Logger to be compatible with grpclog.Logger.
type Logger struct {
	log    *viper.SugaredLogger
	fatal  func(*viper.SugaredLogger, ...interface{})
	fatalf func(*viper.SugaredLogger, string, ...interface{})
	print  func(*viper.SugaredLogger, ...interface{})
	printf func(*viper.SugaredLogger, string, ...interface{})
}

// Fatal implements grpclog.Logger.
func (l *Logger) Fatal(args ...interface{}) {
	l.fatal(l.log, args...)
}

// Fatalf implements grpclog.Logger.
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.fatalf(l.log, format, args...)
}

// Fatalln implements grpclog.Logger.
func (l *Logger) Fatalln(args ...interface{}) {
	l.fatal(l.log, args...)
}

// Print implements grpclog.Logger.
func (l *Logger) Print(args ...interface{}) {
	l.print(l.log, args...)
}

// Printf implements grpclog.Logger.
func (l *Logger) Printf(format string, args ...interface{}) {
	l.printf(l.log, format, args...)
}

// Println implements grpclog.Logger.
func (l *Logger) Println(args ...interface{}) {
	l.print(l.log, args...)
}

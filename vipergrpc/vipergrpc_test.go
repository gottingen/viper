

package vipergrpc

import (
	"testing"

	"github.com/gottingen/viper"
	"github.com/gottingen/viper/vipercore"
	"github.com/gottingen/viper/vipertest/observer"

	"github.com/stretchr/testify/require"
)

func TestLoggerInfoExpected(t *testing.T) {
	checkMessages(t, vipercore.DebugLevel, nil, vipercore.InfoLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugExpected(t *testing.T) {
	checkMessages(t, vipercore.DebugLevel, []Option{WithDebug()}, vipercore.DebugLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerDebugSuppressed(t *testing.T) {
	checkMessages(t, vipercore.InfoLevel, []Option{WithDebug()}, vipercore.DebugLevel, nil, func(logger *Logger) {
		logger.Print("hello")
		logger.Printf("world")
		logger.Println("foo")
	})
}

func TestLoggerFatalExpected(t *testing.T) {
	checkMessages(t, vipercore.DebugLevel, nil, vipercore.FatalLevel, []string{
		"hello",
		"world",
		"foo",
	}, func(logger *Logger) {
		logger.Fatal("hello")
		logger.Fatalf("world")
		logger.Fatalln("foo")
	})
}

func checkMessages(
	t testing.TB,
	enab vipercore.LevelEnabler,
	opts []Option,
	expectedLevel vipercore.Level,
	expectedMessages []string,
	f func(*Logger),
) {
	if expectedLevel == vipercore.FatalLevel {
		expectedLevel = vipercore.WarnLevel
	}
	withLogger(enab, opts, func(logger *Logger, observedLogs *observer.ObservedLogs) {
		f(logger)
		logEntries := observedLogs.All()
		require.Equal(t, len(expectedMessages), len(logEntries))
		for i, logEntry := range logEntries {
			require.Equal(t, expectedLevel, logEntry.Level)
			require.Equal(t, expectedMessages[i], logEntry.Message)
		}
	})
}

func withLogger(
	enab vipercore.LevelEnabler,
	opts []Option,
	f func(*Logger, *observer.ObservedLogs),
) {
	core, observedLogs := observer.New(enab)
	f(NewLogger(viper.New(core), append(opts, withWarn())...), observedLogs)
}

// withWarn redirects the fatal level to the warn level, which makes testing
// easier.
func withWarn() Option {
	return optionFunc(func(logger *Logger) {
		logger.fatal = (*viper.SugaredLogger).Warn
		logger.fatalf = (*viper.SugaredLogger).Warnf
	})
}

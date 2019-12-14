

package viper_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/gottingen/viper"
	"github.com/gottingen/viper/vipercore"
)

func Example_presets() {
	// Using viper's preset constructors is the simplest way to get a feel for the
	// package, but they don't allow much customization.
	logger := viper.NewExample() // or NewProduction, or NewDevelopment
	defer logger.Sync()

	const url = "http://example.com"

	// In most circumstances, use the SugaredLogger. It's 4-10x faster than most
	// other structured logging packages and has a familiar, loosely-typed API.
	sugar := logger.Sugar()
	sugar.Infow("Failed to fetch URL.",
		// Structured context as loosely typed key-value pairs.
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
	sugar.Infof("Failed to fetch URL: %s", url)

	// In the unusual situations where every microsecond matters, use the
	// Logger. It's even faster than the SugaredLogger, but only supports
	// structured logging.
	logger.Info("Failed to fetch URL.",
		// Structured context as strongly typed fields.
		viper.String("url", url),
		viper.Int("attempt", 3),
		viper.Duration("backoff", time.Second),
	)
	// Output:
	// {"level":"info","msg":"Failed to fetch URL.","url":"http://example.com","attempt":3,"backoff":"1s"}
	// {"level":"info","msg":"Failed to fetch URL: http://example.com"}
	// {"level":"info","msg":"Failed to fetch URL.","url":"http://example.com","attempt":3,"backoff":"1s"}
}

func Example_basicConfiguration() {
	// For some users, the presets offered by the NewProduction, NewDevelopment,
	// and NewExample constructors won't be appropriate. For most of those
	// users, the bundled Config struct offers the right balance of flexibility
	// and convenience. (For more complex needs, see the AdvancedConfiguration
	// example.)
	//
	// See the documentation for Config and vipercore.EncoderConfig for all the
	// available options.
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout", "/tmp/logs"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"foo": "bar"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg viper.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("logger construction succeeded")
	// Output:
	// {"level":"info","message":"logger construction succeeded","foo":"bar"}
}

func Example_advancedConfiguration() {
	// The bundled Config struct only supports the most common configuration
	// options. More complex needs, like splitting logs between multiple files
	// or writing to non-file outputs, require use of the vipercore package.
	//
	// In this example, imagine we're both sending our logs to Kafka and writing
	// them to the console. We'd like to encode the console output and the Kafka
	// topics differently, and we'd also like special treatment for
	// high-priority logs.

	// First, define our level-handling logic.
	highPriority := viper.LevelEnablerFunc(func(lvl vipercore.Level) bool {
		return lvl >= vipercore.ErrorLevel
	})
	lowPriority := viper.LevelEnablerFunc(func(lvl vipercore.Level) bool {
		return lvl < vipercore.ErrorLevel
	})

	// Assume that we have clients for two Kafka topics. The clients implement
	// vipercore.WriteSyncer and are safe for concurrent use. (If they only
	// implement io.Writer, we can use vipercore.AddSync to add a no-op Sync
	// method. If they're not safe for concurrent use, we can add a protecting
	// mutex with vipercore.Lock.)
	topicDebugging := vipercore.AddSync(ioutil.Discard)
	topicErrors := vipercore.AddSync(ioutil.Discard)

	// High-priority output should also go to standard error, and low-priority
	// output should also go to standard out.
	consoleDebugging := vipercore.Lock(os.Stdout)
	consoleErrors := vipercore.Lock(os.Stderr)

	// Optimize the Kafka output for machine consumption and the console output
	// for human operators.
	kafkaEncoder := vipercore.NewJSONEncoder(viper.NewProductionEncoderConfig())
	consoleEncoder := vipercore.NewConsoleEncoder(viper.NewDevelopmentEncoderConfig())

	// Join the outputs, encoders, and level-handling functions into
	// vipercore.Cores, then tee the four cores together.
	core := vipercore.NewTee(
		vipercore.NewCore(kafkaEncoder, topicErrors, highPriority),
		vipercore.NewCore(consoleEncoder, consoleErrors, highPriority),
		vipercore.NewCore(kafkaEncoder, topicDebugging, lowPriority),
		vipercore.NewCore(consoleEncoder, consoleDebugging, lowPriority),
	)

	// From a vipercore.Core, it's easy to construct a Logger.
	logger := viper.New(core)
	defer logger.Sync()
	logger.Info("constructed a logger")
}

func ExampleNamespace() {
	logger := viper.NewExample()
	defer logger.Sync()

	logger.With(
		viper.Namespace("metrics"),
		viper.Int("counter", 1),
	).Info("tracked some metrics")
	// Output:
	// {"level":"info","msg":"tracked some metrics","metrics":{"counter":1}}
}

func ExampleNewStdLog() {
	logger := viper.NewExample()
	defer logger.Sync()

	std := viper.NewStdLog(logger)
	std.Print("standard logger wrapper")
	// Output:
	// {"level":"info","msg":"standard logger wrapper"}
}

func ExampleRedirectStdLog() {
	logger := viper.NewExample()
	defer logger.Sync()

	undo := viper.RedirectStdLog(logger)
	defer undo()

	log.Print("redirected standard library")
	// Output:
	// {"level":"info","msg":"redirected standard library"}
}

func ExampleReplaceGlobals() {
	logger := viper.NewExample()
	defer logger.Sync()

	undo := viper.ReplaceGlobals(logger)
	defer undo()

	viper.L().Info("replaced viper's global loggers")
	// Output:
	// {"level":"info","msg":"replaced viper's global loggers"}
}

func ExampleAtomicLevel() {
	atom := viper.NewAtomicLevel()

	// To keep the example deterministic, disable timestamps in the output.
	encoderCfg := viper.NewProductionEncoderConfig()
	encoderCfg.TimeKey = ""

	logger := viper.New(vipercore.NewCore(
		vipercore.NewJSONEncoder(encoderCfg),
		vipercore.Lock(os.Stdout),
		atom,
	))
	defer logger.Sync()

	logger.Info("info logging enabled")

	atom.SetLevel(viper.ErrorLevel)
	logger.Info("info logging disabled")
	// Output:
	// {"level":"info","msg":"info logging enabled"}
}

func ExampleAtomicLevel_config() {
	// The viper.Config struct includes an AtomicLevel. To use it, keep a
	// reference to the Config.
	rawJSON := []byte(`{
		"level": "info",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"encoding": "json",
		"encoderConfig": {
			"messageKey": "message",
			"levelKey": "level",
			"levelEncoder": "lowercase"
		}
	}`)
	var cfg viper.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("info logging enabled")

	cfg.Level.SetLevel(viper.ErrorLevel)
	logger.Info("info logging disabled")
	// Output:
	// {"level":"info","message":"info logging enabled"}
}

func ExampleLogger_Check() {
	logger := viper.NewExample()
	defer logger.Sync()

	if ce := logger.Check(viper.DebugLevel, "debugging"); ce != nil {
		// If debug-level log output isn't enabled or if viper's sampling would have
		// dropped this log entry, we don't allocate the slice that holds these
		// fields.
		ce.Write(
			viper.String("foo", "bar"),
			viper.String("baz", "quux"),
		)
	}

	// Output:
	// {"level":"debug","msg":"debugging","foo":"bar","baz":"quux"}
}

func ExampleLogger_Named() {
	logger := viper.NewExample()
	defer logger.Sync()

	// By default, Loggers are unnamed.
	logger.Info("no name")

	// The first call to Named sets the Logger name.
	main := logger.Named("main")
	main.Info("main logger")

	// Additional calls to Named create a period-separated path.
	main.Named("subpackage").Info("sub-logger")
	// Output:
	// {"level":"info","msg":"no name"}
	// {"level":"info","logger":"main","msg":"main logger"}
	// {"level":"info","logger":"main.subpackage","msg":"sub-logger"}
}

func ExampleWrapCore_replace() {
	// Replacing a Logger's core can alter fundamental behaviors.
	// For example, it can convert a Logger to a no-op.
	nop := viper.WrapCore(func(vipercore.Core) vipercore.Core {
		return vipercore.NewNopCore()
	})

	logger := viper.NewExample()
	defer logger.Sync()

	logger.Info("working")
	logger.WithOptions(nop).Info("no-op")
	logger.Info("original logger still works")
	// Output:
	// {"level":"info","msg":"working"}
	// {"level":"info","msg":"original logger still works"}
}

func ExampleWrapCore_wrap() {
	// Wrapping a Logger's core can extend its functionality. As a trivial
	// example, it can double-write all logs.
	doubled := viper.WrapCore(func(c vipercore.Core) vipercore.Core {
		return vipercore.NewTee(c, c)
	})

	logger := viper.NewExample()
	defer logger.Sync()

	logger.Info("single")
	logger.WithOptions(doubled).Info("doubled")
	// Output:
	// {"level":"info","msg":"single"}
	// {"level":"info","msg":"doubled"}
	// {"level":"info","msg":"doubled"}
}

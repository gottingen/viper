

package benchmarks

import (
	"io/ioutil"

	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
)

func newDisabledApexLog() *log.Logger {
	return &log.Logger{
		Handler: json.New(ioutil.Discard),
		Level:   log.ErrorLevel,
	}
}

func newApexLog() *log.Logger {
	return &log.Logger{
		Handler: json.New(ioutil.Discard),
		Level:   log.DebugLevel,
	}
}

func fakeApexFields() log.Fields {
	return log.Fields{
		"int":     _tenInts[0],
		"ints":    _tenInts,
		"string":  _tenStrings[0],
		"strings": _tenStrings,
		"time":    _tenTimes[0],
		"times":   _tenTimes,
		"user1":   _oneUser,
		"user2":   _oneUser,
		"users":   _tenUsers,
		"error":   errExample,
	}
}

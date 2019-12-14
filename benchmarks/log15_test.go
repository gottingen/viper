

package benchmarks

import (
	"io/ioutil"

	"gopkg.in/inconshreveable/log15.v2"
)

func newLog15() log15.Logger {
	logger := log15.New()
	logger.SetHandler(log15.StreamHandler(ioutil.Discard, log15.JsonFormat()))
	return logger
}

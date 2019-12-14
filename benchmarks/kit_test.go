

package benchmarks

import (
	"io/ioutil"

	"github.com/go-kit/kit/log"
)

func newKitLog(fields ...interface{}) log.Logger {
	return log.With(log.NewJSONLogger(ioutil.Discard), fields...)
}



package benchmarks

import (
	"io/ioutil"

	"github.com/rs/zerolog"
)

func newZerolog() zerolog.Logger {
	return zerolog.New(ioutil.Discard).With().Timestamp().Logger()
}

func newDisabledZerolog() zerolog.Logger {
	return newZerolog().Level(zerolog.Disabled)
}

func fakeZerologFields(e *zerolog.Event) *zerolog.Event {
	return e.
		Int("int", _tenInts[0]).
		Interface("ints", _tenInts).
		Str("string", _tenStrings[0]).
		Interface("strings", _tenStrings).
		Time("time", _tenTimes[0]).
		Interface("times", _tenTimes).
		Interface("user1", _oneUser).
		Interface("user2", _oneUser).
		Interface("users", _tenUsers).
		Err(errExample)
}

func fakeZerologContext(c zerolog.Context) zerolog.Context {
	return c.
		Int("int", _tenInts[0]).
		Interface("ints", _tenInts).
		Str("string", _tenStrings[0]).
		Interface("strings", _tenStrings).
		Time("time", _tenTimes[0]).
		Interface("times", _tenTimes).
		Interface("user1", _oneUser).
		Interface("user2", _oneUser).
		Interface("users", _tenUsers).
		Err(errExample)
}

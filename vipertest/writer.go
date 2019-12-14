

package vipertest

import "github.com/gottingen/viper/internal/vtest"

type (
	// A Syncer is a spy for the Sync portion of vipercore.WriteSyncer.
	Syncer = vtest.Syncer

	// A Discarder sends all writes to ioutil.Discard.
	Discarder = vtest.Discarder

	// FailWriter is a WriteSyncer that always returns an error on writes.
	FailWriter = vtest.FailWriter

	// ShortWriter is a WriteSyncer whose write method never returns an error,
	// but always reports that it wrote one byte less than the input slice's
	// length (thus, a "short write").
	ShortWriter = vtest.ShortWriter

	// Buffer is an implementation of vipercore.WriteSyncer that sends all writes to
	// a bytes.Buffer. It has convenience methods to split the accumulated buffer
	// on newlines.
	Buffer = vtest.Buffer
)

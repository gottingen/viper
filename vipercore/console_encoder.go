

package vipercore

import (
	"fmt"
	"sync"

	"github.com/gottingen/buffer"
)

var _sliceEncoderPool = sync.Pool{
	New: func() interface{} {
		return &sliceArrayEncoder{elems: make([]interface{}, 0, 2)}
	},
}

func getSliceEncoder() *sliceArrayEncoder {
	return _sliceEncoderPool.Get().(*sliceArrayEncoder)
}

func putSliceEncoder(e *sliceArrayEncoder) {
	e.elems = e.elems[:0]
	_sliceEncoderPool.Put(e)
}

type consoleEncoder struct {
	*jsonEncoder
}

// NewConsoleEncoder creates an encoder whose output is designed for human -
// rather than machine - consumption. It serializes the core log entry data
// (message, level, timestamp, etc.) in a plain-text format and leaves the
// structured context as JSON.
//
// Note that although the console encoder doesn't use the keys specified in the
// encoder configuration, it will omit any element whose key is set to the empty
// string.
func NewConsoleEncoder(cfg EncoderConfig) Encoder {
	return consoleEncoder{newJSONEncoder(cfg, true)}
}

func (c consoleEncoder) Clone() Encoder {
	return consoleEncoder{c.jsonEncoder.Clone().(*jsonEncoder)}
}

func (c consoleEncoder) EncodeEntry(ent Entry, fields []Field) (*buffer.Buffer, error) {
	line := buffer.Get()

	// We don't want the entry's metadata to be quoted and escaped (if it's
	// encoded as strings), which means that we can't use the JSON encoder. The
	// simplest option is to use the memory encoder and fmt.Fprint.
	//
	// If this ever becomes a performance bottleneck, we can implement
	// ArrayEncoder for our plain-text format.
	arr := getSliceEncoder()
	if c.TimeKey != "" && c.EncodeTime != nil {
		c.EncodeTime(ent.Time, arr)
	}
	if c.LevelKey != "" && c.EncodeLevel != nil {
		c.EncodeLevel(ent.Level, arr)
	}
	if ent.LoggerName != "" && c.NameKey != "" {
		nameEncoder := c.EncodeName

		if nameEncoder == nil {
			// Fall back to FullNameEncoder for backward compatibility.
			nameEncoder = FullNameEncoder
		}

		nameEncoder(ent.LoggerName, arr)
	}
	if ent.Caller.Defined && c.CallerKey != "" && c.EncodeCaller != nil {
		c.EncodeCaller(ent.Caller, arr)
	}
	for i := range arr.elems {
		if i > 0 {
			line.WriteByte('\t')
		}
		fmt.Fprint(line, arr.elems[i])
	}
	putSliceEncoder(arr)

	// Add the message itself.
	if c.MessageKey != "" {
		c.addTabIfNecessary(line)
		line.WriteString(ent.Message)
	}

	// Add any structured context.
	c.writeContext(line, fields)

	// If there's no stacktrace key, honor that; this allows users to force
	// single-line output.
	if ent.Stack != "" && c.StacktraceKey != "" {
		line.WriteByte('\n')
		line.WriteString(ent.Stack)
	}

	if c.LineEnding != "" {
		line.WriteString(c.LineEnding)
	} else {
		line.WriteString(DefaultLineEnding)
	}
	return line, nil
}

func (c consoleEncoder) writeContext(line *buffer.Buffer, extra []Field) {
	context := c.jsonEncoder.Clone().(*jsonEncoder)
	defer buffer.Put(context.buf)

	addFields(context, extra)
	context.closeOpenNamespaces()
	if context.buf.Len() == 0 {
		return
	}

	c.addTabIfNecessary(line)
	line.WriteByte('{')
	line.Write(context.buf.Bytes())
	line.WriteByte('}')
}

func (c consoleEncoder) addTabIfNecessary(line *buffer.Buffer) {
	if line.Len() > 0 {
		line.WriteByte('\t')
	}
}

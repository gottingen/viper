

package viper

import (
	"runtime"
	"strings"
	"sync"
	"github.com/gottingen/buffer"
)

const _viperPackage = "github.com/gottingen/viper"

var (
	_stacktracePool = sync.Pool{
		New: func() interface{} {
			return newProgramCounters(64)
		},
	}

	// We add "." and "/" suffixes to the package name to ensure we only match
	// the exact package and not any package with the same prefix.
	_viperStacktracePrefixes       = addPrefix(_viperPackage, ".", "/")
	_viperStacktraceVendorContains = addPrefix("/vendor/", _viperStacktracePrefixes...)
)

func takeStacktrace() string {
	buf := buffer.Get()
	defer buffer.Put(buf)
	programCounters := _stacktracePool.Get().(*programCounters)
	defer _stacktracePool.Put(programCounters)

	var numFrames int
	for {
		// Skip the call to runtime.Counters and takeStacktrace so that the
		// program counters start at the caller of takeStacktrace.
		numFrames = runtime.Callers(2, programCounters.pcs)
		if numFrames < len(programCounters.pcs) {
			break
		}
		// Don't put the too-short counter slice back into the pool; this lets
		// the pool adjust if we consistently take deep stacktraces.
		programCounters = newProgramCounters(len(programCounters.pcs) * 2)
	}

	i := 0
	skipViperFrames := true // skip all consecutive viper frames at the beginning.
	frames := runtime.CallersFrames(programCounters.pcs[:numFrames])

	// Note: On the last iteration, frames.Next() returns false, with a valid
	// frame, but we ignore this frame. The last frame is a a runtime frame which
	// adds noise, since it's only either runtime.main or runtime.goexit.
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		if skipViperFrames && isViperFrame(frame.Function) {
			continue
		} else {
			skipViperFrames = false
		}

		if i != 0 {
			buf.WriteByte('\n')
		}
		i++
		buf.WriteString(frame.Function)
		buf.WriteByte('\n')
		buf.WriteByte('\t')
		buf.WriteString(frame.File)
		buf.WriteByte(':')
		buf.WriteInt(int64(frame.Line))
	}

	return buf.String()
}

func isViperFrame(function string) bool {
	for _, prefix := range _viperStacktracePrefixes {
		if strings.HasPrefix(function, prefix) {
			return true
		}
	}

	// We can't use a prefix match here since the location of the vendor
	// directory affects the prefix. Instead we do a contains match.
	for _, contains := range _viperStacktraceVendorContains {
		if strings.Contains(function, contains) {
			return true
		}
	}

	return false
}

type programCounters struct {
	pcs []uintptr
}

func newProgramCounters(size int) *programCounters {
	return &programCounters{make([]uintptr, size)}
}

func addPrefix(prefix string, ss ...string) []string {
	withPrefix := make([]string, len(ss))
	for i, s := range ss {
		withPrefix[i] = prefix + s
	}
	return withPrefix
}

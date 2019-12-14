

// Package observer provides a vipercore.Core that keeps an in-memory,
// encoding-agnostic repesentation of log entries. It's useful for
// applications that want to unit test their log output without tying their
// tests to a particular output encoding.
package observer // import "github.com/gottingen/viper/vipertest/observer"

import (
	"strings"
	"sync"
	"time"

	"github.com/gottingen/viper/vipercore"
)

// ObservedLogs is a concurrency-safe, ordered collection of observed logs.
type ObservedLogs struct {
	mu   sync.RWMutex
	logs []LoggedEntry
}

// Len returns the number of items in the collection.
func (o *ObservedLogs) Len() int {
	o.mu.RLock()
	n := len(o.logs)
	o.mu.RUnlock()
	return n
}

// All returns a copy of all the observed logs.
func (o *ObservedLogs) All() []LoggedEntry {
	o.mu.RLock()
	ret := make([]LoggedEntry, len(o.logs))
	for i := range o.logs {
		ret[i] = o.logs[i]
	}
	o.mu.RUnlock()
	return ret
}

// TakeAll returns a copy of all the observed logs, and truncates the observed
// slice.
func (o *ObservedLogs) TakeAll() []LoggedEntry {
	o.mu.Lock()
	ret := o.logs
	o.logs = nil
	o.mu.Unlock()
	return ret
}

// AllUntimed returns a copy of all the observed logs, but overwrites the
// observed timestamps with time.Time's zero value. This is useful when making
// assertions in tests.
func (o *ObservedLogs) AllUntimed() []LoggedEntry {
	ret := o.All()
	for i := range ret {
		ret[i].Time = time.Time{}
	}
	return ret
}

// FilterMessage filters entries to those that have the specified message.
func (o *ObservedLogs) FilterMessage(msg string) *ObservedLogs {
	return o.filter(func(e LoggedEntry) bool {
		return e.Message == msg
	})
}

// FilterMessageSnippet filters entries to those that have a message containing the specified snippet.
func (o *ObservedLogs) FilterMessageSnippet(snippet string) *ObservedLogs {
	return o.filter(func(e LoggedEntry) bool {
		return strings.Contains(e.Message, snippet)
	})
}

// FilterField filters entries to those that have the specified field.
func (o *ObservedLogs) FilterField(field vipercore.Field) *ObservedLogs {
	return o.filter(func(e LoggedEntry) bool {
		for _, ctxField := range e.Context {
			if ctxField.Equals(field) {
				return true
			}
		}
		return false
	})
}

func (o *ObservedLogs) filter(match func(LoggedEntry) bool) *ObservedLogs {
	o.mu.RLock()
	defer o.mu.RUnlock()

	var filtered []LoggedEntry
	for _, entry := range o.logs {
		if match(entry) {
			filtered = append(filtered, entry)
		}
	}
	return &ObservedLogs{logs: filtered}
}

func (o *ObservedLogs) add(log LoggedEntry) {
	o.mu.Lock()
	o.logs = append(o.logs, log)
	o.mu.Unlock()
}

// New creates a new Core that buffers logs in memory (without any encoding).
// It's particularly useful in tests.
func New(enab vipercore.LevelEnabler) (vipercore.Core, *ObservedLogs) {
	ol := &ObservedLogs{}
	return &contextObserver{
		LevelEnabler: enab,
		logs:         ol,
	}, ol
}

type contextObserver struct {
	vipercore.LevelEnabler
	logs    *ObservedLogs
	context []vipercore.Field
}

func (co *contextObserver) Check(ent vipercore.Entry, ce *vipercore.CheckedEntry) *vipercore.CheckedEntry {
	if co.Enabled(ent.Level) {
		return ce.AddCore(ent, co)
	}
	return ce
}

func (co *contextObserver) With(fields []vipercore.Field) vipercore.Core {
	return &contextObserver{
		LevelEnabler: co.LevelEnabler,
		logs:         co.logs,
		context:      append(co.context[:len(co.context):len(co.context)], fields...),
	}
}

func (co *contextObserver) Write(ent vipercore.Entry, fields []vipercore.Field) error {
	all := make([]vipercore.Field, 0, len(fields)+len(co.context))
	all = append(all, co.context...)
	all = append(all, fields...)
	co.logs.add(LoggedEntry{ent, all})
	return nil
}

func (co *contextObserver) Sync() error {
	return nil
}

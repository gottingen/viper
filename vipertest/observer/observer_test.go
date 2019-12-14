

package observer_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gottingen/viper"
	"github.com/gottingen/viper/vipercore"
	. "github.com/gottingen/viper/vipertest/observer"
)

func assertEmpty(t testing.TB, logs *ObservedLogs) {
	assert.Equal(t, 0, logs.Len(), "Expected empty ObservedLogs to have zero length.")
	assert.Equal(t, []LoggedEntry{}, logs.All(), "Unexpected LoggedEntries in empty ObservedLogs.")
}

func TestObserver(t *testing.T) {
	observer, logs := New(viper.InfoLevel)
	assertEmpty(t, logs)

	assert.NoError(t, observer.Sync(), "Unexpected failure in no-op Sync")

	obs := viper.New(observer).With(viper.Int("i", 1))
	obs.Info("foo")
	obs.Debug("bar")
	want := []LoggedEntry{{
		Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "foo"},
		Context: []vipercore.Field{viper.Int("i", 1)},
	}}

	assert.Equal(t, 1, logs.Len(), "Unexpected observed logs Len.")
	assert.Equal(t, want, logs.AllUntimed(), "Unexpected contents from AllUntimed.")

	all := logs.All()
	require.Equal(t, 1, len(all), "Unexpected numbed of LoggedEntries returned from All.")
	assert.NotEqual(t, time.Time{}, all[0].Time, "Expected non-zero time on LoggedEntry.")

	// copy & zero time for stable assertions
	untimed := append([]LoggedEntry{}, all...)
	untimed[0].Time = time.Time{}
	assert.Equal(t, want, untimed, "Unexpected LoggedEntries from All.")

	assert.Equal(t, all, logs.TakeAll(), "Expected All and TakeAll to return identical results.")
	assertEmpty(t, logs)
}

func TestObserverWith(t *testing.T) {
	sf1, logs := New(viper.InfoLevel)

	// need to pad out enough initial fields so that the underlying slice cap()
	// gets ahead of its len() so that the sf3/4 With append's could choose
	// not to copy (if the implementation doesn't force them)
	sf1 = sf1.With([]vipercore.Field{viper.Int("a", 1), viper.Int("b", 2)})

	sf2 := sf1.With([]vipercore.Field{viper.Int("c", 3)})
	sf3 := sf2.With([]vipercore.Field{viper.Int("d", 4)})
	sf4 := sf2.With([]vipercore.Field{viper.Int("e", 5)})
	ent := vipercore.Entry{Level: viper.InfoLevel, Message: "hello"}

	for i, core := range []vipercore.Core{sf2, sf3, sf4} {
		if ce := core.Check(ent, nil); ce != nil {
			ce.Write(viper.Int("i", i))
		}
	}

	assert.Equal(t, []LoggedEntry{
		{
			Entry: ent,
			Context: []vipercore.Field{
				viper.Int("a", 1),
				viper.Int("b", 2),
				viper.Int("c", 3),
				viper.Int("i", 0),
			},
		},
		{
			Entry: ent,
			Context: []vipercore.Field{
				viper.Int("a", 1),
				viper.Int("b", 2),
				viper.Int("c", 3),
				viper.Int("d", 4),
				viper.Int("i", 1),
			},
		},
		{
			Entry: ent,
			Context: []vipercore.Field{
				viper.Int("a", 1),
				viper.Int("b", 2),
				viper.Int("c", 3),
				viper.Int("e", 5),
				viper.Int("i", 2),
			},
		},
	}, logs.All(), "expected no field sharing between With siblings")
}

func TestFilters(t *testing.T) {
	logs := []LoggedEntry{
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "log a"},
			Context: []vipercore.Field{viper.String("fStr", "1"), viper.Int("a", 1)},
		},
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "log a"},
			Context: []vipercore.Field{viper.String("fStr", "2"), viper.Int("b", 2)},
		},
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "log b"},
			Context: []vipercore.Field{viper.Int("a", 1), viper.Int("b", 2)},
		},
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "log c"},
			Context: []vipercore.Field{viper.Int("a", 1), viper.Namespace("ns"), viper.Int("a", 2)},
		},
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "msg 1"},
			Context: []vipercore.Field{viper.Int("a", 1), viper.Namespace("ns")},
		},
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "any map"},
			Context: []vipercore.Field{viper.Any("map", map[string]string{"a": "b"})},
		},
		{
			Entry:   vipercore.Entry{Level: viper.InfoLevel, Message: "any slice"},
			Context: []vipercore.Field{viper.Any("slice", []string{"a"})},
		},
	}

	logger, sink := New(viper.InfoLevel)
	for _, log := range logs {
		logger.Write(log.Entry, log.Context)
	}

	tests := []struct {
		msg      string
		filtered *ObservedLogs
		want     []LoggedEntry
	}{
		{
			msg:      "filter by message",
			filtered: sink.FilterMessage("log a"),
			want:     logs[0:2],
		},
		{
			msg:      "filter by field",
			filtered: sink.FilterField(viper.String("fStr", "1")),
			want:     logs[0:1],
		},
		{
			msg:      "filter by message and field",
			filtered: sink.FilterMessage("log a").FilterField(viper.Int("b", 2)),
			want:     logs[1:2],
		},
		{
			msg:      "filter by field with duplicate fields",
			filtered: sink.FilterField(viper.Int("a", 2)),
			want:     logs[3:4],
		},
		{
			msg:      "filter doesn't match any messages",
			filtered: sink.FilterMessage("no match"),
			want:     []LoggedEntry{},
		},
		{
			msg:      "filter by snippet",
			filtered: sink.FilterMessageSnippet("log"),
			want:     logs[0:4],
		},
		{
			msg:      "filter by snippet and field",
			filtered: sink.FilterMessageSnippet("a").FilterField(viper.Int("b", 2)),
			want:     logs[1:2],
		},
		{
			msg:      "filter for map",
			filtered: sink.FilterField(viper.Any("map", map[string]string{"a": "b"})),
			want:     logs[5:6],
		},
		{
			msg:      "filter for slice",
			filtered: sink.FilterField(viper.Any("slice", []string{"a"})),
			want:     logs[6:7],
		},
	}

	for _, tt := range tests {
		got := tt.filtered.AllUntimed()
		assert.Equal(t, tt.want, got, tt.msg)
	}
}

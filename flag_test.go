

package viper

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/gottingen/viper/vipercore"

	"github.com/stretchr/testify/assert"
)

type flagTestCase struct {
	args      []string
	wantLevel vipercore.Level
	wantErr   bool
}

func (tc flagTestCase) runImplicitSet(t testing.TB) {
	origCommandLine := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flag.CommandLine.SetOutput(ioutil.Discard)
	defer func() { flag.CommandLine = origCommandLine }()

	level := LevelFlag("level", InfoLevel, "")
	tc.run(t, flag.CommandLine, level)
}

func (tc flagTestCase) runExplicitSet(t testing.TB) {
	var lvl vipercore.Level
	set := flag.NewFlagSet("test", flag.ContinueOnError)
	set.Var(&lvl, "level", "minimum enabled logging level")
	tc.run(t, set, &lvl)
}

func (tc flagTestCase) run(t testing.TB, set *flag.FlagSet, actual *vipercore.Level) {
	err := set.Parse(tc.args)
	if tc.wantErr {
		assert.Error(t, err, "Parse(%v) should fail.", tc.args)
		return
	}
	if assert.NoError(t, err, "Parse(%v) should succeed.", tc.args) {
		assert.Equal(t, tc.wantLevel, *actual, "Level mismatch.")
	}
}

func TestLevelFlag(t *testing.T) {
	tests := []flagTestCase{
		{
			args:      nil,
			wantLevel: vipercore.InfoLevel,
		},
		{
			args:    []string{"--level", "unknown"},
			wantErr: true,
		},
		{
			args:      []string{"--level", "error"},
			wantLevel: vipercore.ErrorLevel,
		},
	}

	for _, tt := range tests {
		tt.runExplicitSet(t)
		tt.runImplicitSet(t)
	}
}

func TestLevelFlagsAreIndependent(t *testing.T) {
	origCommandLine := flag.CommandLine
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flag.CommandLine.SetOutput(ioutil.Discard)
	defer func() { flag.CommandLine = origCommandLine }()

	// Make sure that these two flags are independent.
	fileLevel := LevelFlag("file-level", InfoLevel, "")
	consoleLevel := LevelFlag("console-level", InfoLevel, "")

	assert.NoError(t, flag.CommandLine.Parse([]string{"-file-level", "debug"}), "Unexpected flag-parsing error.")
	assert.Equal(t, InfoLevel, *consoleLevel, "Expected file logging level to remain unchanged.")
	assert.Equal(t, DebugLevel, *fileLevel, "Expected console logging level to have changed.")
}

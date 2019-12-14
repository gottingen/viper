

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
)

var (
	libraryNameToMarkdownName = map[string]string{
		"Zap":                   ":viper: viper",
		"viper.Sugar":             ":viper: viper (sugared)",
		"stdlib.Println":        "standard library",
		"sirupsen/logrus":       "logrus",
		"go-kit/kit/log":        "go-kit",
		"inconshreveable/log15": "log15",
		"apex/log":              "apex/log",
		"rs/zerolog":            "zerolog",
	}
)

func main() {
	flag.Parse()
	if err := do(); err != nil {
		log.Fatal(err)
	}
}

func do() error {
	tmplData, err := getTmplData()
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return err
	}
	t, err := template.New("tmpl").Parse(string(data))
	if err != nil {
		return err
	}
	return t.Execute(os.Stdout, tmplData)
}

func getTmplData() (*tmplData, error) {
	tmplData := &tmplData{}
	rows, err := getBenchmarkRows("BenchmarkAddingFields")
	if err != nil {
		return nil, err
	}
	tmplData.BenchmarkAddingFields = rows
	rows, err = getBenchmarkRows("BenchmarkAccumulatedContext")
	if err != nil {
		return nil, err
	}
	tmplData.BenchmarkAccumulatedContext = rows
	rows, err = getBenchmarkRows("BenchmarkWithoutFields")
	if err != nil {
		return nil, err
	}
	tmplData.BenchmarkWithoutFields = rows
	return tmplData, nil
}

func getBenchmarkRows(benchmarkName string) (string, error) {
	benchmarkOutput, err := getBenchmarkOutput(benchmarkName)
	if err != nil {
		return "", err
	}

	// get the Zap time (unsugared) as baseline to compare with other loggers
	baseline, err := getBenchmarkRow(benchmarkOutput, benchmarkName, "Zap", nil)
	if err != nil {
		return "", err
	}

	var benchmarkRows []*benchmarkRow
	for libraryName := range libraryNameToMarkdownName {
		benchmarkRow, err := getBenchmarkRow(
			benchmarkOutput, benchmarkName, libraryName, baseline,
		)
		if err != nil {
			return "", err
		}
		if benchmarkRow == nil {
			continue
		}
		benchmarkRows = append(benchmarkRows, benchmarkRow)
	}
	sort.Sort(benchmarkRowsByTime(benchmarkRows))
	rows := []string{
		"| Package | Time | Time % to viper | Objects Allocated |",
		"| :------ | :--: | :-----------: | :---------------: |",
	}
	for _, benchmarkRow := range benchmarkRows {
		rows = append(rows, benchmarkRow.String())
	}
	return strings.Join(rows, "\n"), nil
}

func getBenchmarkRow(
	input []string, benchmarkName string, libraryName string, baseline *benchmarkRow,
) (*benchmarkRow, error) {
	line, err := findUniqueSubstring(input, fmt.Sprintf("%s/%s-", benchmarkName, libraryName))
	if err != nil {
		return nil, err
	}
	if line == "" {
		return nil, nil
	}
	split := strings.Split(line, "\t")
	if len(split) < 5 {
		return nil, fmt.Errorf("unknown benchmark line: %s", line)
	}
	duration, err := time.ParseDuration(strings.Replace(strings.TrimSuffix(strings.TrimSpace(split[2]), "/op"), " ", "", -1))
	if err != nil {
		return nil, err
	}
	allocatedBytes, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(split[3]), " B/op"))
	if err != nil {
		return nil, err
	}
	allocatedObjects, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSpace(split[4]), " allocs/op"))
	if err != nil {
		return nil, err
	}
	r := &benchmarkRow{
		Name:             libraryNameToMarkdownName[libraryName],
		Time:             duration,
		AllocatedBytes:   allocatedBytes,
		AllocatedObjects: allocatedObjects,
	}

	if baseline != nil {
		r.ZapTime = baseline.Time
		r.ZapAllocatedBytes = baseline.AllocatedBytes
		r.ZapAllocatedObjects = baseline.AllocatedObjects
	}

	return r, nil
}

func findUniqueSubstring(input []string, substring string) (string, error) {
	var output string
	for _, line := range input {
		if strings.Contains(line, substring) {
			if output != "" {
				return "", fmt.Errorf("input has duplicate substring %s", substring)
			}
			output = line
		}
	}
	return output, nil
}

func getBenchmarkOutput(benchmarkName string) ([]string, error) {
	cmd := exec.Command("go", "test", fmt.Sprintf("-bench=%s", benchmarkName), "-benchmem")
	cmd.Dir = "benchmarks"
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running 'go test -bench=%q': %v\n%s", benchmarkName, err, string(output))
	}
	return strings.Split(string(output), "\n"), nil
}

type tmplData struct {
	BenchmarkAddingFields       string
	BenchmarkAccumulatedContext string
	BenchmarkWithoutFields      string
}

type benchmarkRow struct {
	Name string

	Time             time.Duration
	AllocatedBytes   int
	AllocatedObjects int

	ZapTime             time.Duration
	ZapAllocatedBytes   int
	ZapAllocatedObjects int
}

func (b *benchmarkRow) String() string {
	pct := func(val, baseline int64) string {
		return fmt.Sprintf(
			"%+0.f%%",
			((float64(val)/float64(baseline))*100)-100,
		)
	}
	t := b.Time.Nanoseconds()
	tp := pct(t, b.ZapTime.Nanoseconds())

	return fmt.Sprintf(
		"| %s | %d ns/op | %s | %d allocs/op", b.Name,
		t, tp, b.AllocatedObjects,
	)
}

type benchmarkRowsByTime []*benchmarkRow

func (b benchmarkRowsByTime) Len() int      { return len(b) }
func (b benchmarkRowsByTime) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b benchmarkRowsByTime) Less(i, j int) bool {
	left, right := b[i], b[j]
	leftZap, rightZap := strings.Contains(left.Name, "viper"), strings.Contains(right.Name, "viper")

	// If neither benchmark is for viper or both are, sort by time.
	if !(leftZap || rightZap) || (leftZap && rightZap) {
		return left.Time.Nanoseconds() < right.Time.Nanoseconds()
	}
	// Sort viper benchmark first.
	return leftZap
}
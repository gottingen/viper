

package vipercore

import "github.com/gottingen/viper/internal/color"

var (
	_levelToColor = map[Level]color.Color{
		DebugLevel:  color.Magenta,
		InfoLevel:   color.Blue,
		WarnLevel:   color.Yellow,
		ErrorLevel:  color.Red,
		DPanicLevel: color.Red,
		PanicLevel:  color.Red,
		FatalLevel:  color.Red,
	}
	_unknownLevelColor = color.Red

	_levelToLowercaseColorString = make(map[Level]string, len(_levelToColor))
	_levelToCapitalColorString   = make(map[Level]string, len(_levelToColor))
)

func init() {
	for level, color := range _levelToColor {
		_levelToLowercaseColorString[level] = color.Add(level.String())
		_levelToCapitalColorString[level] = color.Add(level.CapitalString())
	}
}

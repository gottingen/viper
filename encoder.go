

package viper

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gottingen/viper/vipercore"
)

var (
	errNoEncoderNameSpecified = errors.New("no encoder name specified")

	_encoderNameToConstructor = map[string]func(vipercore.EncoderConfig) (vipercore.Encoder, error){
		"console": func(encoderConfig vipercore.EncoderConfig) (vipercore.Encoder, error) {
			return vipercore.NewConsoleEncoder(encoderConfig), nil
		},
		"json": func(encoderConfig vipercore.EncoderConfig) (vipercore.Encoder, error) {
			return vipercore.NewJSONEncoder(encoderConfig), nil
		},
	}
	_encoderMutex sync.RWMutex
)

// RegisterEncoder registers an encoder constructor, which the Config struct
// can then reference. By default, the "json" and "console" encoders are
// registered.
//
// Attempting to register an encoder whose name is already taken returns an
// error.
func RegisterEncoder(name string, constructor func(vipercore.EncoderConfig) (vipercore.Encoder, error)) error {
	_encoderMutex.Lock()
	defer _encoderMutex.Unlock()
	if name == "" {
		return errNoEncoderNameSpecified
	}
	if _, ok := _encoderNameToConstructor[name]; ok {
		return fmt.Errorf("encoder already registered for name %q", name)
	}
	_encoderNameToConstructor[name] = constructor
	return nil
}

func newEncoder(name string, encoderConfig vipercore.EncoderConfig) (vipercore.Encoder, error) {
	_encoderMutex.RLock()
	defer _encoderMutex.RUnlock()
	if name == "" {
		return nil, errNoEncoderNameSpecified
	}
	constructor, ok := _encoderNameToConstructor[name]
	if !ok {
		return nil, fmt.Errorf("no encoder registered for name %q", name)
	}
	return constructor(encoderConfig)
}

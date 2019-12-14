

package viper

import (
	"time"

	"github.com/gottingen/viper/vipercore"
)

// Array constructs a field with the given key and ArrayMarshaler. It provides
// a flexible, but still type-safe and efficient, way to add array-like types
// to the logging context. The struct's MarshalLogArray method is called lazily.
func Array(key string, val vipercore.ArrayMarshaler) Field {
	return Field{Key: key, Type: vipercore.ArrayMarshalerType, Interface: val}
}

// Bools constructs a field that carries a slice of bools.
func Bools(key string, bs []bool) Field {
	return Array(key, bools(bs))
}

// ByteStrings constructs a field that carries a slice of []byte, each of which
// must be UTF-8 encoded text.
func ByteStrings(key string, bss [][]byte) Field {
	return Array(key, byteStringsArray(bss))
}

// Complex128s constructs a field that carries a slice of complex numbers.
func Complex128s(key string, nums []complex128) Field {
	return Array(key, complex128s(nums))
}

// Complex64s constructs a field that carries a slice of complex numbers.
func Complex64s(key string, nums []complex64) Field {
	return Array(key, complex64s(nums))
}

// Durations constructs a field that carries a slice of time.Durations.
func Durations(key string, ds []time.Duration) Field {
	return Array(key, durations(ds))
}

// Float64s constructs a field that carries a slice of floats.
func Float64s(key string, nums []float64) Field {
	return Array(key, float64s(nums))
}

// Float32s constructs a field that carries a slice of floats.
func Float32s(key string, nums []float32) Field {
	return Array(key, float32s(nums))
}

// Ints constructs a field that carries a slice of integers.
func Ints(key string, nums []int) Field {
	return Array(key, ints(nums))
}

// Int64s constructs a field that carries a slice of integers.
func Int64s(key string, nums []int64) Field {
	return Array(key, int64s(nums))
}

// Int32s constructs a field that carries a slice of integers.
func Int32s(key string, nums []int32) Field {
	return Array(key, int32s(nums))
}

// Int16s constructs a field that carries a slice of integers.
func Int16s(key string, nums []int16) Field {
	return Array(key, int16s(nums))
}

// Int8s constructs a field that carries a slice of integers.
func Int8s(key string, nums []int8) Field {
	return Array(key, int8s(nums))
}

// Strings constructs a field that carries a slice of strings.
func Strings(key string, ss []string) Field {
	return Array(key, stringArray(ss))
}

// Times constructs a field that carries a slice of time.Times.
func Times(key string, ts []time.Time) Field {
	return Array(key, times(ts))
}

// Uints constructs a field that carries a slice of unsigned integers.
func Uints(key string, nums []uint) Field {
	return Array(key, uints(nums))
}

// Uint64s constructs a field that carries a slice of unsigned integers.
func Uint64s(key string, nums []uint64) Field {
	return Array(key, uint64s(nums))
}

// Uint32s constructs a field that carries a slice of unsigned integers.
func Uint32s(key string, nums []uint32) Field {
	return Array(key, uint32s(nums))
}

// Uint16s constructs a field that carries a slice of unsigned integers.
func Uint16s(key string, nums []uint16) Field {
	return Array(key, uint16s(nums))
}

// Uint8s constructs a field that carries a slice of unsigned integers.
func Uint8s(key string, nums []uint8) Field {
	return Array(key, uint8s(nums))
}

// Uintptrs constructs a field that carries a slice of pointer addresses.
func Uintptrs(key string, us []uintptr) Field {
	return Array(key, uintptrs(us))
}

// Errors constructs a field that carries a slice of errors.
func Errors(key string, errs []error) Field {
	return Array(key, errArray(errs))
}

type bools []bool

func (bs bools) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range bs {
		arr.WriteBool(bs[i])
	}
	return nil
}

type byteStringsArray [][]byte

func (bss byteStringsArray) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range bss {
		arr.WriteByteString(bss[i])
	}
	return nil
}

type complex128s []complex128

func (nums complex128s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendComplex128(nums[i])
	}
	return nil
}

type complex64s []complex64

func (nums complex64s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.AppendComplex64(nums[i])
	}
	return nil
}

type durations []time.Duration

func (ds durations) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range ds {
		arr.AppendDuration(ds[i])
	}
	return nil
}

type float64s []float64

func (nums float64s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteFloat64(nums[i])
	}
	return nil
}

type float32s []float32

func (nums float32s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteFloat32(nums[i])
	}
	return nil
}

type ints []int

func (nums ints) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteInt(nums[i])
	}
	return nil
}

type int64s []int64

func (nums int64s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteInt64(nums[i])
	}
	return nil
}

type int32s []int32

func (nums int32s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteInt32(nums[i])
	}
	return nil
}

type int16s []int16

func (nums int16s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteInt16(nums[i])
	}
	return nil
}

type int8s []int8

func (nums int8s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteInt8(nums[i])
	}
	return nil
}

type stringArray []string

func (ss stringArray) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range ss {
		arr.WriteString(ss[i])
	}
	return nil
}

type times []time.Time

func (ts times) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range ts {
		arr.AppendTime(ts[i])
	}
	return nil
}

type uints []uint

func (nums uints) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteUint(nums[i])
	}
	return nil
}

type uint64s []uint64

func (nums uint64s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteUint64(nums[i])
	}
	return nil
}

type uint32s []uint32

func (nums uint32s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteUint32(nums[i])
	}
	return nil
}

type uint16s []uint16

func (nums uint16s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteUint16(nums[i])
	}
	return nil
}

type uint8s []uint8

func (nums uint8s) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteUint8(nums[i])
	}
	return nil
}

type uintptrs []uintptr

func (nums uintptrs) MarshalLogArray(arr vipercore.ArrayEncoder) error {
	for i := range nums {
		arr.WriteUintptr(nums[i])
	}
	return nil
}

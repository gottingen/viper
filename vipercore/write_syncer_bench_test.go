

package vipercore

import (
	"testing"

	"github.com/gottingen/viper/internal/vtest"
)

func BenchmarkMultiWriteSyncer(b *testing.B) {
	b.Run("2", func(b *testing.B) {
		w := NewMultiWriteSyncer(
			&vtest.Discarder{},
			&vtest.Discarder{},
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w.Write([]byte("foobarbazbabble"))
			}
		})
	})
	b.Run("4", func(b *testing.B) {
		w := NewMultiWriteSyncer(
			&vtest.Discarder{},
			&vtest.Discarder{},
			&vtest.Discarder{},
			&vtest.Discarder{},
		)
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				w.Write([]byte("foobarbazbabble"))
			}
		})
	})
}

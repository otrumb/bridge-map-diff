package snapshot_test

import (
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func FuzzParse_never_panics(f *testing.F) {
	f.Add([]byte(`{"version":1,"mappings":[]}`))
	f.Add([]byte(`{"version":2}`))
	f.Add([]byte{0xff, 0x00})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = snapshot.Parse(data)
	})
}

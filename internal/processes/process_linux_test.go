//go:build linux

package processes

import (
	"encoding/binary"
	"testing"
)

func TestClockTicksPerSecond(t *testing.T) {
	hz := clockTicksPerSecond()
	if hz <= 0 {
		t.Fatalf("clockTicksPerSecond() = %d, want positive", hz)
	}
	if hz != 100 {
		t.Logf("clockTicksPerSecond() = %d (unusual, but non-zero)", hz)
	}
}

func TestNativeUint(t *testing.T) {
	if binary.NativeEndian.Uint16([]byte{0x01, 0x00}) != 1 {
		t.Skip("little-endian only")
	}
	cases := []struct {
		data []byte
		word int
		want uint64
	}{
		{[]byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 8, 1},
		{[]byte{0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, 8, 100},
		{[]byte{0x11, 0x00, 0x00, 0x00}, 4, 17},
	}
	for _, c := range cases {
		if got := nativeUint(c.data, c.word); got != c.want {
			t.Errorf("nativeUint(%v, %d) = %d, want %d", c.data, c.word, got, c.want)
		}
	}
}

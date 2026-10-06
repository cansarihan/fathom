package sample

import "testing"

// FuzzParseRecords drives the fragile parser. The engine finds a length that
// overruns the frame quickly and records the failing input.
func FuzzParseRecords(f *testing.F) {
	f.Add([]byte{0x01, 0x00, 0x02, 'h', 'i'})
	f.Add([]byte{0x00})
	f.Fuzz(func(t *testing.T, frame []byte) {
		_ = ParseRecords(frame)
	})
}

// FuzzChecksum exercises a total function: it should never fail, so it gives
// the campaign a target that only grows coverage.
func FuzzChecksum(f *testing.F) {
	f.Add([]byte("seed"))
	f.Fuzz(func(t *testing.T, data []byte) {
		if Checksum(data) == 0 && len(data) > 0 {
			// 0 is reachable; this is a property check, not a real fault.
			_ = data
		}
	})
}

// Package sample is a small, deliberately fragile wire format used to
// demonstrate Fathom. ParseRecords has a bounds bug that a fuzzer finds in
// well under a second; Checksum and the round trip are there to give a campaign
// more than one target to schedule.
package sample

import "encoding/binary"

// ParseRecords walks a length-prefixed frame:
//
//	[1 byte count][ count × [2 byte big-endian length][length payload bytes] ]
//
// It trusts the declared lengths — a crafted length runs the slice past the end
// of the frame and panics, which is the fault the demo campaign uncovers.
func ParseRecords(frame []byte) [][]byte {
	if len(frame) == 0 {
		return nil
	}
	count := int(frame[0])
	offset := 1
	out := make([][]byte, 0, count)
	for i := 0; i < count; i++ {
		length := int(binary.BigEndian.Uint16(frame[offset : offset+2]))
		offset += 2
		payload := frame[offset : offset+length]
		out = append(out, payload)
		offset += length
	}
	return out
}

// Checksum folds a frame into a 32-bit value. It is total over all inputs.
func Checksum(data []byte) uint32 {
	var sum uint32 = 2166136261
	for _, b := range data {
		sum ^= uint32(b)
		sum *= 16777619
	}
	return sum
}

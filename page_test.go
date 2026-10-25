package goberkeleydb

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

func TestPageHeaderSize(t *testing.T) {
	if got := unsafe.Sizeof(pageHeader{}); got != headerSize {
		t.Errorf("unsafe.Sizeof(pageHeader{}) = %d, want %d", got, headerSize)
	}
	// unsafe.Sizeof includes padding the compiler inserts to align fields,
	// so it stays at headerSize even if an explicit padding field is
	// removed. encoding/binary writes fields back to back without that
	// padding, so binary.Size catches it.
	if got := binary.Size(pageHeader{}); got != headerSize {
		t.Errorf("binary.Size(pageHeader{}) = %d, want %d", got, headerSize)
	}
}

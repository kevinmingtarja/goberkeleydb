package goberkeleydb

import "encoding/binary"

// dbMetaSize is the number of bytes at the start of page 0 that hold the
// meta fields and are covered by its checksum. Opening a file reads this
// many bytes before the page size is known, so it must not exceed the
// smallest page size.
const dbMetaSize = 512

// LSN is a byte position in the log. It only ever increases.
type LSN uint64

// Meta page layout (page 0):
//
//	+-----------------------------------+
//	|    lsn    |   pgno    |   magic   |
//	+-----------------------------------+
//	|  version  | pagesize  |  (pad)    |
//	+-----------------------------------+
//	|   type    |  (pad)    |  chksum   |
//	+-----------------------------------+
//	|   root    |   free    | last pgno |
//	+-----------------------------------+
//	|            U N U S E D            |
//	+-----------------------------------+ <- dbMetaSize
//	|            U N U S E D            |
//	+-----------------------------------+ <- page size
//
// lsn, pgno, type and chksum are at the same offsets as in pageHeader, so
// code that handles any page can read them without knowing the page type.
// The fields end at byte 44 and the rest of the page is unused. chksum
// covers the first dbMetaSize bytes, so page 0 can be verified before the
// page size is known.
type metaPage struct {
	LSN      LSN     // 00-07: Log sequence number.
	Pgno     uint32  // 08-11: Current page number.
	Magic    uint32  // 12-15: Identifies the file format.
	Version  uint32  // 16-19: Format version.
	PageSize uint32  // 20-23: Page size in bytes.
	_        byte    // 24: Padding.
	Type     uint8   // 25: Page type.
	_        [2]byte // 26-27: Padding.
	Chksum   uint32  // 28-31: CRC32C of the first dbMetaSize bytes.
	Root     uint32  // 32-35: Root page number (always 1).
	Free     uint32  // 36-39: Head of the free page list.
	LastPgno uint32  // 40-43: Last page number in the file.
}

// Page layout:
//
//	+-----------------------------------+
//	|    lsn    |   pgno    | prev pgno |
//	+-----------------------------------+
//	| next pgno |  entries  |  hoffset  |
//	+-----------------------------------+
//	|   level   |   type    |  (pad)    |
//	+-----------------------------------+
//	|  chksum   |   index   | free -->  |
//	+-----------+-----------------------+
//	|          F R E E A R E A          |
//	+-----------------------------------+
//	|              <-- free |   item    |
//	+-----------------------------------+
//	|   item    |   item    |   item    |
//	+-----------------------------------+
//
// The header is 32 bytes and the index (a uint16 offset per item) always
// starts at byte 32. Items are allocated from the end of the page downward;
// hoffset is the offset of the lowest item, so the free area is
// [32 + 2*entries, hoffset).
type pageHeader struct {
	LSN     LSN     // 00-07: Log sequence number.
	Pgno    uint32  // 08-11: Current page number.
	Prev    uint32  // 12-15: Previous page number.
	Next    uint32  // 16-19: Next page number.
	Entries uint16  // 20-21: Number of items on the page.
	HOffset uint16  // 22-23: High free byte page offset.
	Level   uint8   // 24: Btree tree level.
	Type    uint8   // 25: Page type.
	_       [2]byte // 26-27: Padding.
	Chksum  uint32  // 28-31: CRC32C of the whole page.
}

const headerSize = 32

type Page []byte

const slotSize = 2 // each slot is a uint16 offset

// slot returns the offset of item i (P_INP(dbp, pg)[i]).
func (p Page) slot(i int) uint16 {
	return binary.LittleEndian.Uint16(p[headerSize+slotSize*i:])
}

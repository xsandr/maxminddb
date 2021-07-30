package maxminddb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

type Type uint

const (
	Extended Type = iota
	Pointer
	String
	Double
	Bytes
	Uint16
	Uint32
	Map
	Int32
	Uint64
	Uint128
	Array
	Container
	EndMarker
	Boolean
	Float
)

func NewDecoder(buffer []byte, initialOffset int) decoder {
	return decoder{Cursor{buffer: buffer, cursor: initialOffset}}
}

type decoder struct {
	Cursor
}

func (d *decoder) decodeDottedMap(fields []string, result map[string]interface{}) error {
	initialCursorOffset := d.cursor
	for _, field := range fields {
		d.cursor = initialCursorOffset
		if err := d.findField(field, []byte(field), result); err != nil {
			return err
		}
	}
	return nil
}

func (d *decoder) findField(field string, parts []byte, result map[string]interface{}) error {
	if len(parts) == 0 {
		result[field] = d.decodeValue()
		return nil
	}

	var searchFor []byte
	isIndex := true

	for i, char := range parts {
		if char == '.' {
			searchFor = parts[:i]
			parts = parts[i+1:]
			break
		}
		if !(isNum(char)) {
			isIndex = false
		}
	}
	// last piece of the path
	if searchFor == nil {
		searchFor = parts
		parts = parts[len(parts):]
	}

	var idx int
	if isIndex {
		idx = byteSliceToInt(searchFor)
	}

	dataType, size := d.decodeControlByte()
	if dataType == Pointer {
		d.cursor = d.getPointerAddress()
		dataType, size = d.decodeControlByte()
	}

	if isIndex {
		if dataType != Array {
			return fmt.Errorf("cannot use indices for the field")
		}
	} else if dataType != Map {
		return fmt.Errorf("expected a map")
	}

	for i := 0; i < size; i++ {
		if isIndex {
			if i == idx {
				return d.findField(field, parts, result)
			}
		} else {
			key := d.decodeStringAsBytes()
			if bytes.Equal(key, searchFor) {
				return d.findField(field, parts, result)
			}
		}
		d.skipValue()
	}
	return nil
}

func (d *decoder) getPointerAddress() int {
	// getPointerAddress might be called only after decodeControlByte
	ctrlByte := d.buffer[d.cursor-1]
	size := d.getSize(ctrlByte)
	pointerByteSize := (size >> 3) & 0x3
	switch pointerByteSize {
	default: // 0
		return ((size & 0x7) << 8) + int(d.currentByte())
	case 1:
		return 2048 + (((size & 0x7) << 16) | int(d.currentByte())<<8 | int(d.currentByte()))
	case 2:
		return 526336 + (((size & 0x7) << 24) | int(d.currentByte())<<16 | int(d.currentByte())<<8 | int(d.currentByte()))
	case 3:
		return int(d.currentByte())<<24 | int(d.currentByte())<<16 | int(d.currentByte())<<8 | int(d.currentByte())
	}
}

func (d *decoder) getSize(ctrlByte uint8) int {
	// last 5 bits represent the size of the data structure
	size := int(ctrlByte & 0x1f)
	// if size < 29 than it's size in bytes, otherwise:
	if size >= 29 {
		if size == 29 {
			size = 29 + int(d.decodeUint(1))
		} else if size == 30 {
			size = 285 + int(d.decodeUint(2))
		} else if size == 31 {
			size = 65821 + int(d.decodeUint(3))
		}
	}
	return size
}

func (d *decoder) decodeControlByte() (Type, int) {
	ctrlByte := d.currentByte()
	// first 3 bits represent type
	t := Type(ctrlByte >> 5)
	if t == Extended {
		// extended - means that the 7 + the values in the next byte represent real type
		t = Type(7 + d.currentByte())
	}
	size := d.getSize(ctrlByte)
	return t, size
}

func (d *decoder) skipValue() {
	valueType, size := d.decodeControlByte()
	switch valueType {
	case Int32, Uint16, Uint32, Uint64, Uint128, String, Bytes:
		d.moveCaret(size)
	case Pointer:
		d.getPointerAddress()
	case Float:
		d.moveCaret(4)
	case Double:
		d.moveCaret(8)
	case Array:
		for i := 0; i < size; i++ {
			d.skipValue()
		}
	case Map:
		for i := 0; i < size; i++ {
			// for map we have to skip both: key and value
			d.skipValue()
			d.skipValue()
		}
	}
}

func (d *decoder) decodeUint(n int) uint {
	bytesToDecode := d.nextBytes(n)
	v := uint(bytesToDecode[0])
	for i:=1;i<len(bytesToDecode);i++{
		v = v<<8 | uint(bytesToDecode[i])
	}
	return v
}

func (d *decoder) decodeStringAsBytes() []byte {
	stype, size := d.decodeControlByte()
	switch stype {
	case String:
		return d.nextBytes(size)
	case Pointer:
		pointerOffset := d.getPointerAddress()
		initial := d.cursor
		d.cursor = pointerOffset + 1
		size := d.getSize(d.buffer[pointerOffset])
		result := d.buffer[d.cursor : d.cursor+size]
		d.cursor = initial
		return result
	default:
		panic("unexpected type")
	}
}

func (d *decoder) decodeString() string {
	stype, size := d.decodeControlByte()
	switch stype {
	case String:
		return BytesToString(d.nextBytes(size))
	case Pointer:
		// resolve pointer right here
		initial := d.cursor
		d.cursor = d.getPointerAddress()
		result := d.decodeString()
		d.cursor = initial + 1
		return result
	default:
		panic("unexpected type for string decoding")
	}
}

func (d *decoder) decodeValue() interface{} {
	dataType, size := d.decodeControlByte()
	switch dataType {
	case String:
		byteSlice := d.nextBytes(size)
		return BytesToString(byteSlice)
	case Uint16, Uint32:
		return d.decodeUint(size)
	case Double:
		u64 := binary.BigEndian.Uint64(d.nextBytes(size))
		return math.Float64frombits(u64)
	case Float:
		u32 := binary.BigEndian.Uint32(d.nextBytes(size))
		return math.Float32frombits(u32)
	case Boolean:
		return d.currentByte() > 0
	case Pointer:
		d.cursor = d.getPointerAddress()
		return d.decodeValue()
	default:
		return nil
	}
}

func byteSliceToInt(b []byte) int {
	var result int
	for i:=0;i<len(b);i++{
		result = result*10 + int(b[i]-'0')
	}
	return result
}

func isNum(char byte) bool {
	return '0' <= char && char <= '9'
}

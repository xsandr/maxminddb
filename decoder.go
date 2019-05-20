package maxmind

import (
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

type decoder struct {
	buffer []byte
	cursor int //cursor points on current position in the buffer
}

func (d *decoder) moveCarret(n int) {
	d.cursor += n
}

func (d *decoder) currentByte() byte {
	d.cursor += 1
	return d.buffer[d.cursor-1]
}

func (d *decoder) nextBytes(n int) []byte {
	d.cursor += n
	return d.buffer[d.cursor-n : d.cursor]
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

func convertToInt(b []byte) int {
	var result int
	for _, char := range b {
		result = result*10 + int(char-'0')
	}
	return result
}

func isNum(char byte) bool {
	return '0' <= char && char <= '9'
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
		idx = convertToInt(searchFor)
	}

	dataType, mapSize := d.decodeControlByte()
	if dataType == Pointer {
		d.cursor = d.getPointerAddress()
		dataType, mapSize = d.decodeControlByte()
	}

	if isIndex {
		if dataType != Array {
			return fmt.Errorf("cannot use indices for the field: %s", field)
		}
	} else if dataType != Map {
		return fmt.Errorf("expected a map for the field: %s", field)
	}

	for i := 0; i < mapSize; i++ {
		if isIndex {
			if i == idx {
				return d.findField(field, parts, result)
			}
		} else {
			key := d.decodeStringAsBytes()
			if equal(key, searchFor) {
				return d.findField(field, parts, result)
			}
		}
		d.skipValue()
	}
	return nil
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, char := range a {
		if char != b[i] {
			return false
		}
	}
	return true
}

func (d *decoder) getPointerAddress() int {
	// assume that we've just called decodeControlByte
	// and realised that we have a pointer here
	ctrlByte := d.buffer[d.cursor-1]
	size := int(ctrlByte & 0x18 >> 3)

	switch size {
	default:
		return int(ctrlByte&0x7)<<8 + int(d.currentByte())
	case 1:
		return 2048 + int(ctrlByte&0x7)<<16 | int(d.currentByte())<<8 | int(d.currentByte())
	case 2:
		return 526336 + int(ctrlByte&0x7)<<24 | int(d.currentByte())<<16 | int(d.currentByte())<<8 | int(d.currentByte())
	case 3:
		return int(d.currentByte())<<24 | int(d.currentByte())<<16 | int(d.currentByte())<<8 | int(d.currentByte())
	}
}

// assumes that offset point to control byte
func (d *decoder) decodeControlByte() (Type, int) {
	ctrlByte := d.currentByte()
	// first 3 bits represent type
	t := Type(ctrlByte >> 5)

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

	if t == Extended {
		// extended means that the next byte contains real type
		t = Type(7 + d.currentByte())
	}
	return t, size
}

func (d *decoder) decodeMap(fields []string, result map[string]interface{}) {
	_, mapSize := d.decodeControlByte()
	for i := 0; i < mapSize; i++ {
		key := d.decodeString()
		if contains(fields, key) {
			result[key] = d.decodeValue()
		} else {
			d.skipValue()
		}
	}
}

func (d *decoder) skipValue() {
	// TODO add other types
	valueType, size := d.decodeControlByte()
	switch valueType {
	case Int32, Uint16, Uint32, Uint64, Uint128, String, Bytes:
		d.moveCarret(size)
	case Pointer:
		d.moveCarret(1)
	case Float:
		d.moveCarret(4)
	case Double:
		d.moveCarret(8)
	case Array:
		for i := 0; i < size; i++ {
			d.skipValue()
		}
	case Map:
		for i := 0; i < size; i++ {
			// skip key and value
			d.skipValue()
			d.skipValue()
		}
	}

}

func (d *decoder) decodeUint(n int) uint {
	bytesToDecode := d.nextBytes(n)
	v := uint(bytesToDecode[0])
	for _, b := range bytesToDecode[1:] {
		v = v<<8 | uint(b)
	}
	return v
}

func (d *decoder) decodeStringAsBytes() []byte {
	stype, size := d.decodeControlByte()
	switch stype {
	case String:
		return d.nextBytes(size)
	case Pointer:
		initial := d.cursor
		d.cursor = d.getPointerAddress()
		result := d.decodeStringAsBytes()
		d.cursor = initial + 1
		return result
	default:
		panic(fmt.Sprintf("Unexpected type %v", stype))
	}
}

func (d *decoder) decodeString() string {
	stype, size := d.decodeControlByte()
	switch stype {
	case String:
		return string(d.nextBytes(size))
	case Pointer:
		// resolve pointer right here
		initial := d.cursor
		d.cursor = d.getPointerAddress()
		result := d.decodeString()
		d.cursor = initial + 1
		return result
	default:
		panic(fmt.Sprintf("Unexpected type %v", stype))
	}
}

func (d *decoder) decodeValue() interface{} {
	dataType, size := d.decodeControlByte()
	switch dataType {
	case String:
		return string(d.nextBytes(size))
	case Uint16, Uint32:
		return d.decodeUint(size)
	case Double:
		u64 := binary.BigEndian.Uint64(d.nextBytes(size))
		return math.Float64frombits(u64)
	case Float:
		u32 := binary.BigEndian.Uint32(d.nextBytes(size))
		return math.Float32frombits(u32)
	case Boolean:
		return uint(d.currentByte()) > 0
	default:
		return nil
	}
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

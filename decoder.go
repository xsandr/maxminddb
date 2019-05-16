package maxmind

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

func (d *decoder) decodeMap(fields []string) map[string]interface{} {
	result := make(map[string]interface{})
	_, mapSize := d.decodeControlByte()
	for i := 0; i < mapSize; i++ {
		key := d.decodeString()
		if contains(fields, key) {
			result[key] = d.decodeValue()
		} else {
			d.skipValue()
		}
	}
	return result
}

func (d *decoder) skipValue() {
	valueType, size := d.decodeControlByte()
	switch valueType {
	case Int32, Uint16, Uint32, Uint64, Uint128, String:
		d.cursor += size
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

func (d *decoder) decodeString() string {
	_, size := d.decodeControlByte()
	return string(d.nextBytes(size))
}

func (d *decoder) decodeValue() interface{} {
	dataType, size := d.decodeControlByte()
	switch dataType {
	case String:
		return string(d.nextBytes(size))
	case Uint16, Uint32:
		return d.decodeUint(size)
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

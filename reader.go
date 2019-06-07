package maxmind

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
)

// Reader provides methods for IP lookups
type Reader struct {
	buffer   []byte
	Metadata *Metadata
}

// Open reads the file and return ready to use Reader instance
func Open(filepath string) (*Reader, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	metadata, err := ParseMetadata(data)
	if err != nil {
		return nil, err
	}
	return &Reader{data, metadata}, err
}

// Lookup tries to find the IP in the buffer and puts requested fields into result map
func (r *Reader) Lookup(ip net.IP, fields []string, result map[string]interface{}) error {
	searchTreeSize := (int(r.Metadata.RecordSize*2) / 8 * int(r.Metadata.NodeCount)) + 16
	ipOffset, err := r.FindIPOffset(ip)
	if err != nil {
		return err
	}
	// ipOffset is a relative to data section, not the beginning of the buffer
	ipOffset = ipOffset - int(r.Metadata.NodeCount) - 16
	d := decoder{r.buffer[searchTreeSize:], ipOffset}
	return d.decodeDottedMap(fields, result)
}

// copy from net package
var v4InV6Prefix = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}

func (r *Reader) FindIPOffset(ipAddr net.IP) (int, error) {
	nodeSizeInByte := int(r.Metadata.RecordSize * 2 / 8)

	bitMask, size := convertIPToBigEndian(ipAddr)

	offset := 0
	if size == 32 {
		offset = 96 * nodeSizeInByte
	}

	var v uint
	for i := size - 1; i >= 0; i-- {
		isLeft := (bitMask>>uint(i))&1 == 0
		node := r.buffer[offset : offset+nodeSizeInByte]

		switch r.Metadata.RecordSize {
		case 28:
			var b []byte
			value := r.buffer[offset+3]
			if isLeft {
				b = node[:3]
				value = value >> 4
			} else {
				value = value & 0x0F
				b = node[4:]
			}
			v = uint(value)
			for _, oneByte := range b {
				v = v<<8 | uint(oneByte)
			}

		default:
			var b []byte
			if isLeft {
				b = node[:nodeSizeInByte/2]
			} else {
				b = node[nodeSizeInByte/2:]
			}
			v := uint(b[0])
			for _, oneByte := range b[1:] {
				v = v<<8 | uint(oneByte)
			}
		}
		if v == r.Metadata.NodeCount {
			return 0, fmt.Errorf("couln't find the ip %s", ipAddr.String())
		} else if v < r.Metadata.NodeCount {
			offset = nodeSizeInByte * int(v)
		} else {
			return int(v), nil
		}
	}
	if v == r.Metadata.NodeCount {
		return 0, fmt.Errorf("couln't find the ip %s", ipAddr.String())
	} else if v < r.Metadata.NodeCount {
		offset = int(r.Metadata.RecordSize * v)
	}
	return int(v), nil
}

func convertIPToBigEndian(ipAddr net.IP) (uint64, int) {
	ip := []byte(ipAddr)
	size := 128
	if bytes.HasPrefix(ip, v4InV6Prefix) {
		ip = ip[len(v4InV6Prefix):]
		size = 32
	}
	v := uint64(ip[0])
	for _, i := range ip[1:] {
		v = v<<8 | uint64(i)
	}
	return v, size

}

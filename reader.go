package maxmind

import (
	"io/ioutil"
	"net"
)

// Reader provides methods for IP lookups
type Reader struct {
	buffer   []byte
	Metadata Metadata
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
	return nil
}

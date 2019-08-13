package maxminddb

import (
	"bytes"
	"fmt"
)

// Metadata stores info about opened maxmind file
type Metadata struct {
	NodeCount  uint
	RecordSize uint
	IPVersion  uint
}

var metadataSeparator = []byte("\xab\xcd\xefMaxMind.com")

// ParseMetadata decodes metadata section of the file
// and returns important parameters
func ParseMetadata(buffer []byte) (*Metadata, error) {
	start := bytes.LastIndex(buffer, metadataSeparator)
	if start == -1 {
		return nil, fmt.Errorf("couldn't find a metadata section separator")
	}

	d := NewDecoder(buffer, start+len(metadataSeparator))

	fieldList := []string{"node_count", "record_size", "ip_version"}
	data := make(map[string]interface{})
	if err := d.decodeDottedMap(fieldList, data); err != nil {
		return nil, err
	}

	metadata := &Metadata{
		NodeCount:  data["node_count"].(uint),
		RecordSize: data["record_size"].(uint),
		IPVersion:  data["ip_version"].(uint),
	}

	return metadata, nil
}

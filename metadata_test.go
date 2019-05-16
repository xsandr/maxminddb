package maxmind

import (
	"io/ioutil"
	"testing"
)

func TestMetadataParsing(t *testing.T) {
	data, err := ioutil.ReadFile("test_data/test-data/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Fatal()
	}
	metadata, err := ParseMetadata(data)
	if err != nil || metadata == nil {
		t.Fatal()
	}
	if metadata.NodeCount != 1431 || metadata.RecordSize != 28 || metadata.IPVersion != 6 {
		t.Fatal()
	}
}

package maxmind

import "testing"

func TestMetadataParsing(t *testing.T) {
	db, err := Open("test_data/test-data/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Fail()
	}
	if db.Metadata.NodeCount != 1200 {
		t.Fail()
	}
}

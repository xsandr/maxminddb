package maxmind

import (
	"net"
	"testing"
)

func TestMaxmindLookup(t *testing.T) {
	ip := net.ParseIP("89.160.20.112")
	result := make(map[string]interface{})
	fields := []string{"country.iso_code"}

	db, err := Open("test_data/test-data/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Error(err)
	}

	if err = db.Lookup(ip, fields, result); err != nil {
		t.Error(err)
	}
	isoCode, ok := result["country.iso_code"]
	if !ok {
		t.Error("couldn't find the country.iso_code in the results map")
	}
	isoCodeString, ok := isoCode.(string)
	if !ok {
		t.Error("couldn't convert isoCode to the string")
	}
	if isoCodeString != "SE" {
		t.Error("isoCodeString != 'SE'")
	}
}

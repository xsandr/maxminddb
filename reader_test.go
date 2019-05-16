package maxmind

import (
	"net"
	"testing"
)

func TestMaxmindLookup(t *testing.T) {
	ip := net.ParseIP("81.2.69.160")
	result := make(map[string]interface{})
	// country iso_code
	fields := []string{"country.iso_code", "country.names.en"}

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

	countryName := result["country.names.en"].(string)
	if isoCodeString != "GB" || countryName != "United Kingdom" {
		t.Error()
	}

}

func BenchmarkMaxmindLookup(b *testing.B) {
	ip := net.ParseIP("81.2.69.160")
	result := make(map[string]interface{})
	fields := []string{"country.iso_code", "country.names.en"}
	db, err := Open("test_data/test-data/GeoIP2-City-Test.mmdb")
	if err != nil || db == nil {
		b.Fail()
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if err = db.Lookup(ip, fields, result); err != nil {
			b.Fail()
		}
		isoCode := result["country.iso_code"].(string)
		countryName := result["country.names.en"].(string)

		if isoCode != "GB" || countryName != "United Kingdom" {
			b.Fail()
		}
	}
}

package maxmind

import (
	"net"
	"testing"
)

func TestMaxmindLookup(t *testing.T) {
	ip := net.ParseIP("81.2.69.160")
	result := make(map[string]interface{})
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

func TestArrayIndices(t *testing.T) {
	ip := net.ParseIP("81.2.69.160")
	result := make(map[string]interface{})
	fields := []string{"subdivisions.0.names.en"}

	db, err := Open("test_data/test-data/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Error(err)
	}

	if err = db.Lookup(ip, fields, result); err != nil {
		t.Error(err)
	}
	regionName, ok := result["subdivisions.0.names.en"]
	if !ok {
		t.Error("couldn't find the region's name")
	}

	if regionName.(string) != "England" {
		t.Error()
	}
}

func TestDouble(t *testing.T) {
	ip := net.ParseIP("81.2.69.160")
	result := make(map[string]interface{})
	fields := []string{"location.latitude", "location.longitude"}

	db, err := Open("test_data/test-data/GeoIP2-City-Test.mmdb")
	if err != nil {
		t.Error(err)
	}

	if err = db.Lookup(ip, fields, result); err != nil {
		t.Error(err)
	}
	if latitude, ok := result["location.latitude"]; !ok {
		t.Error("couldn't find latitude")
	} else {
		if latitude.(float64) != 51.5142 {
			t.Fail()
		}
	}
	if longitude, ok := result["location.longitude"]; !ok {
		t.Error("couldn't find longitude")
	} else {
		if longitude.(float64) != -0.0931 {
			t.Fail()
		}
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

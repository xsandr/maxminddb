#Maxminddb

The package provides a reader that can make a lookup into maxmind db file and fetch only requested fields

Example

```go
	db, err := maxminddb.Open("path/to/GeoIP2-City.mmdb")
	if err != nil {
		t.Error(err)
	}

	ip := net.ParseIP("216.160.83.56")
	result := make(map[string]interface{})
	fields := []string{
		"location.metro_code",
		"subdivisions.0.names.en"
	}
	if err = db.Lookup(ip, fields, result); err != nil {
		log.Error("couldn't get the record due to %s", err)
	}
```
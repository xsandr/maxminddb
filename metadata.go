package maxmind

// Metadata stores info about opened maxmind file
type Metadata struct {
	NodeCount  int
	RecordSize int
	IpVersion  int
}

func ParseMetadata(buffer []byte) (Metadata, error) {
	
}

package l5x

import (
	"encoding/xml"
	"io"
	"os"
)

func ParseFromFile(path string) (*RSLogix5000Content, error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	content, err := NewFromReader(xmlFile)
	if err != nil {
		return nil, err
	}

	return content, nil
}

// NewFromReader parses a reader which provides data in RSLogix5000 L5X format.
func NewFromReader(rd io.Reader) (*RSLogix5000Content, error) {
	dec := xml.NewDecoder(rd)

	content := &RSLogix5000Content{}
	err := dec.Decode(content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

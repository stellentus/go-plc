package l5x

import (
	"encoding/xml"
	"io/ioutil"
	"os"
)

func ParseFromFile(path string) (*RSLogix5000Content, error) {
	xmlFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer xmlFile.Close()

	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		return nil, err
	}

	content := &RSLogix5000Content{}

	err = xml.Unmarshal(byteValue, content)
	if err != nil {
		return nil, err
	}

	return content, nil
}

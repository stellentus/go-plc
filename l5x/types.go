package l5x

import (
	"encoding/xml"
	"strings"
	"time"
)

type RSLogix5000Content struct {
	XMLName          xml.Name    `xml:"RSLogix5000Content"`
	SchemaRevision   float32     `xml:",attr"`
	SoftwareRevision float32     `xml:",attr"`
	TargetName       string      `xml:",attr"`
	TargetType       string      `xml:",attr"`
	ContainsContext  bool        `xml:",attr"`
	ExportDate       rsLogixTime `xml:"ExportDate,attr"`
	ExportOptions    stringSlice `xml:"ExportOptions,attr"`
	Controller
}

type Controller struct {
}

type rsLogixTime time.Time

func (rlt *rsLogixTime) fromString(str string) error {
	const layout = "Mon Jan 2 15:04:05 2006"
	parse, err := time.Parse(layout, str)
	if err != nil {
		return err
	}
	*rlt = rsLogixTime(parse)
	return nil
}

func (rlt *rsLogixTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	d.DecodeElement(&str, &start)
	return rlt.fromString(str)
}

func (rlt *rsLogixTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return rlt.fromString(attr.Value)
}

// stringSlice is imported from a string of space-separated strings
type stringSlice struct{ strings []string }

func (ss *stringSlice) fromString(str string) error {
	for _, st := range strings.Split(str, " ") {
		ss.strings = append(ss.strings, st)
	}
	return nil
}

func (ss *stringSlice) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	d.DecodeElement(&str, &start)
	return ss.fromString(str)
}

func (ss *stringSlice) UnmarshalXMLAttr(attr xml.Attr) error {
	return ss.fromString(attr.Value)
}

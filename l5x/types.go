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
	Controller       Controller
}

type Controller struct {
	Use                      string      `xml:",attr"`
	Name                     string      `xml:",attr"`
	ProcessorType            string      `xml:",attr"`
	MajorRev                 int         `xml:",attr"`
	MinorRev                 int         `xml:",attr"`
	TimeSlice                int         `xml:",attr"`
	ShareUnusedTimeSlice     int         `xml:",attr"`
	ProjectCreationDate      rsLogixTime `xml:",attr"`
	LastModifiedDate         rsLogixTime `xml:",attr"`
	SFCExecutionControl      string      `xml:",attr"`
	SFCRestartPosition       string      `xml:",attr"`
	SFCLastScan              string      `xml:",attr"`
	ProjectSN                string      `xml:",attr"`
	MatchProjectToController bool        `xml:",attr"`
	CanUseRPIFromProducer    bool        `xml:",attr"`
	BlockAutoUpdate          bool        `xml:"InhibitAutomaticFirmwareUpdate,attr"`
	PassThroughConfiguration string      `xml:",attr"`
	DownloadProjDocs         bool        `xml:"DownloadProjectDocumentationAndExtendedProperties,attr"`
	DownloadProjProps        bool        `xml:"DownloadProjectCustomProperties,attr"`
	ReportMinorOverflow      bool        `xml:",attr"`
	RedundancyInfo           RedundancyInfo
	Security                 Security
	SecurityInfo             SecurityInfo
	DataTypes                []DataType `xml:"DataTypes>DataType"`
	Modules                  Modules
	AddOnInstructions        AddOnInstructions `xml:"AddOnInstructionDefinitions"`
	Tags                     Tags
	Programs                 Programs
	Tasks                    Tasks
	CST                      CST
	WallClockTime            WallClockTime
	Trends                   Trends
	DataLogs                 DataLogs
	TimeSynchronize          TimeSynchronize
	EthernetPorts            EthernetPorts
	EthernetNetwork          EthernetNetwork
}

type RedundancyInfo struct {
	Enabled                   bool `xml:",attr"`
	KeepTestEditsOnSwitchOver bool `xml:",attr"`
	IOMemoryPadPercentage     int  `xml:",attr"`
	DataTablePadPercentage    int  `xml:",attr"`
}

type Security struct {
	Code            int    `xml:",attr"`
	ChangesToDetect string `xml:",attr"`
}

type SecurityInfo struct{}

type DataType struct {
	Name    string   `xml:",attr"`
	Family  string   `xml:",attr"`
	Class   string   `xml:",attr"` // TODO: enum
	Members []Member `xml:"Members>Member"`
}

type Member struct {
	Name           string `xml:",attr"`
	DataType       string `xml:",attr"` // TODO: enum
	Dimension      int    `xml:",attr"`
	Radix          string `xml:",attr"` // TODO: enum
	Hidden         bool   `xml:",attr"`
	BitNumber      int         `xml:",attr,omitempty"`
	ExternalAccess string `xml:",attr"` // TODO: enum
}

type Modules struct {
}

type AddOnInstructions struct {
}

type Tags struct {
}

type Programs struct {
}

type Tasks struct {
}

type CST struct {
	MasterID int `xml:",attr"`
}

type WallClockTime struct {
	LocalTimeAdjustment int `xml:",attr"`
	TimeZone            int `xml:",attr"`
}

type Trends struct {
}

type DataLogs struct{}

type TimeSynchronize struct {
	Priority1 int  `xml:",attr"`
	Priority2 int  `xml:",attr"`
	PTPEnable bool `xml:",attr"`
}

type EthernetPorts struct {
}

type EthernetNetwork struct {
	SupervisorModeEnabled bool `xml:",attr"`
	SupervisorPrecedence  int  `xml:",attr"`
	BeaconInterval        int  `xml:",attr"`
	BeaconTimeout         int  `xml:",attr"`
	VLANID                int  `xml:",attr"`
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

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
	SecurityInfo             struct{}
	DataTypes                []DataType        `xml:"DataTypes>DataType"`
	Modules                  []Module          `xml:"Modules>Module"`
	AddOnInstructions        AddOnInstructions `xml:"AddOnInstructionDefinitions"`
	Tags                     Tags
	Programs                 Programs
	Tasks                    Tasks
	CST                      CST
	WallClockTime            WallClockTime
	Trends                   Trends
	DataLogs                 struct{}
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
	BitNumber      int    `xml:",attr,omitempty"`
	ExternalAccess string `xml:",attr"` // TODO: enum
	Description    Description `xml:",omitempty"`
}

type Description struct {
		Cdata string `xml:",cdata"`
}

type Module struct {
	Name            string `xml:",attr"`
	CatalogNumber   string `xml:",attr"`
	Vendor          int    `xml:",attr"`
	ProductType     int    `xml:",attr"`
	ProductCode     int    `xml:",attr"`
	Major           int    `xml:",attr"`
	Minor           int    `xml:",attr"`
	ParentModule    string `xml:",attr"`
	ParentModPortId int    `xml:",attr"`
	Inhibited       bool   `xml:",attr"`
	MajorFault      bool   `xml:",attr"`
	EKey            struct {
		State string `xml:",attr"`
	}
	Ports              []Port `xml:"Ports>Port"`
	Communications     Communications
	ExtendedProperties ExtendedProperties
}

type Port struct {
	Id       int    `xml:",attr"`
	Address  int    `xml:",attr,omitempty"`
	Type     string `xml:",attr"` // TODO: enum
	Upstream bool   `xml:",attr"`
	Bus      struct {
		Size int `xml:",attr,omitempty"`
	}
}

type Communications struct {
	ConfigTag   ConfigTag
	Connections []Connection `xml:"Connections>Connection"`
}

type ConfigTag struct {
	ConfigSize     int    `xml:",attr"`
	ExternalAccess string `xml:",attr"` // TODO: enum
	Data           []Data
}

type Data struct {
	Format    string    `xml:",attr"`  // TODO: enum
	L5K       string    `xml:",cdata"` // TODO: would be nice to omitempty
	Structure Structure `xml:",omitempty"`
}

type Structure struct {
	DataType        string `xml:",attr"`
	DataValueMember []DataValueMember
}

type DataValueMember struct {
	Name     string `xml:",attr"`
	DataType string `xml:",attr"` // TODO: enum
	Radix    string `xml:",attr"` // TODO: enum
	Value    string `xml:",attr"`
}

type Connection struct {
	Name        string `xml:",attr"`
	RPI         int    `xml:",attr"`
	Type        string `xml:",attr"`
	EventID     int    `xml:",attr"`
	SendTrigger bool   `xml:"ProgrammaticallySendEventTrigger,attr"`
	InputTag    IOTag
	OutputTag   IOTag
}

type IOTag struct {
	ExternalAccess string    `xml:",attr"`
	Comments       []Comment `xml:"Comments>Comment,omitempty"`
	Data           []Data
}

type Comment struct {
	Operand string `xml:",attr"`
	Cdata   string `xml:",cdata"`
}

type ExtendedProperties struct {
	Public struct {
		ConfigID int
		CatNum   string
	} `xml:"public"`
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

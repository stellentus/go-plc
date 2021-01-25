package l5x

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// NewFromFile parses the RSLogix5000 L5X file at the provided path.
func NewFromFile(path string) (*RSLogix5000Content, error) {
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
	DataTypes                []DataType      `xml:"DataTypes>DataType"`
	Modules                  []Module        `xml:"Modules>Module"`
	AddOnInstrDefs           []AddOnInstrDef `xml:"AddOnInstructionDefinitions>AddOnInstructionDefinition"`
	Tags                     TagList         `xml:"Tags>Tag"`
	Programs                 []Program       `xml:"Programs>Program"`
	Tasks                    []Task          `xml:"Tasks>Task"`
	CST                      CST
	WallClockTime            WallClockTime
	Trends                   []Trend `xml:"Trends>Trend"`
	DataLogs                 struct{}
	TimeSynchronize          TimeSynchronize
	EthernetPorts            []EthernetPort `xml:"EthernetPorts>EthernetPort"`
	EthernetNetwork          EthernetNetwork
}

func (ctrl Controller) TypeList() (TypeList, error) {
	tl := NewTypeList()
	err := tl.AddControlLogixTypes()
	if err != nil {
		return nil, err
	}

	// This loop isn't the most elegant code, but it's fast enough and it's
	// reasonably readable. We loop through the types, creating each one.
	// However, if a type requires a type we haven't yet processed (i.e. we
	// get ErrUnknownType), then we skip it for now. We keep looping through
	// the types until we've created all of them.
	typesToParse := ctrl.DataTypes
	foundAtLeastOne := true // We keep looping as long as the last iteration found one type
	for len(typesToParse) > 0 && foundAtLeastOne {
		foundAtLeastOne = false
		nextLoopMustParse := make([]DataType, 0, len(typesToParse))
		for _, dt := range typesToParse {
			var typ Type
			typ, err = dt.AsType(tl)
			if err != nil {
				if !errors.Is(err, ErrUnknownType) {
					// Whatever err type this is, it's not one that will go away
					return TypeList{}, err
				}
				nextLoopMustParse = append(nextLoopMustParse, dt)
				continue
			}
			tl = append(tl, typ)
			foundAtLeastOne = true
		}
		typesToParse = nextLoopMustParse
	}

	if len(typesToParse) > 0 {
		return TypeList{}, err
	}

	// Add-on instruction definitions work similarly to DataTypes, so we'll treat them the same
	aoTypes, err := ctrl.addOnsAsTypes()
	if err != nil {
		return nil, err
	}
	tl = append(tl, aoTypes...)

	return tl, nil
}

func (ctrl Controller) addOnsAsTypes() (TypeList, error) {
	tl := TypeList{}
	defaultTypes := NewTypeList() // I think add-on instructions can't use custom types?
	for _, ao := range ctrl.AddOnInstrDefs {
		typ, err := ao.AsType(defaultTypes)
		if err != nil {
			return nil, err
		}
		tl = append(tl, typ)
	}
	return tl, nil
}

func (tags TagList) NamedTypes(tl TypeList) ([]NamedType, error) {
	nts := []NamedType{}
	for _, tag := range tags {
		nt, err := tl.newNamedType(tag.Name, tag.DataType, tag.Dimensions)
		if err != nil {
			return nil, fmt.Errorf("Tag '%s' couldn't be created because %w", tag.Name, err)
		}
		nts = append(nts, nt)
	}

	return nts, nil
}

func (prog Program) typesAsStruct(tl TypeList) (structType, error) {
	nts, err := prog.Tags.NamedTypes(tl)
	if err != nil {
		return structType{}, err
	}
	return newStructType(prog.Name, nts)
}

func (ctrl Controller) programStructs(tl TypeList) ([]structType, error) {
	strs := []structType{}

	for _, prog := range ctrl.Programs {
		str, err := prog.typesAsStruct(tl)
		if err != nil {
			return nil, err
		}
		strs = append(strs, str)
	}

	return strs, nil
}

func (ctrl Controller) WriteTagsStruct(tl TypeList, wr io.Writer) error {
	nts, err := ctrl.Tags.NamedTypes(tl)
	if err != nil {
		return err
	}

	pstrs, err := ctrl.programStructs(tl)
	if err != nil {
		return err
	}
	for i, pstr := range pstrs {
		nt, err := newNamedType(ctrl.Programs[i].Name, pstr)
		if err != nil {
			return err
		}
		nt.SetAsProgram()
		nts = append(nts, nt)
		_, err = wr.Write([]byte(TypeDefinition(pstr) + "\n\n"))
	}

	strct, err := newStructType(ctrl.Name, nts)
	if err != nil {
		return err
	}
	_, err = wr.Write([]byte(TypeDefinition(strct) + "\n"))
	return err
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
	Name    string         `xml:",attr"`
	Family  DataTypeFamily `xml:",attr"`
	Class   Class          `xml:",attr"`
	Members []Member       `xml:"Members>Member"`
}

type Member struct {
	Name           string         `xml:",attr"`
	DataType       string         `xml:",attr"` // TODO: enum
	Dimension      int            `xml:",attr"`
	Radix          Radix          `xml:",attr"`
	Hidden         bool           `xml:",attr"`
	BitNumber      int            `xml:",attr,omitempty"`
	Target         string         `xml:",attr,omitempty"` // TODO: must match another Member's name
	ExternalAccess ExternalAccess `xml:",attr"`
	Description    Description    `xml:",omitempty"`
}

type Description struct {
	Cdata string `xml:",cdata"`
}

type Module struct {
	Name               string `xml:",attr"`
	CatalogNumber      string `xml:",attr"`
	Vendor             int    `xml:",attr"`
	ProductType        int    `xml:",attr"`
	ProductCode        int    `xml:",attr"`
	Major              int    `xml:",attr"`
	Minor              int    `xml:",attr"`
	ParentModule       string `xml:",attr"` // TODO: must match another module name
	ParentModPortId    int    `xml:",attr"`
	Inhibited          bool   `xml:",attr"`
	MajorFault         bool   `xml:",attr"`
	EKey               EKeyState_s
	Ports              []Port `xml:"Ports>Port"`
	Communications     Communications
	ExtendedProperties ExtendedProperties
}

type EKeyState_s struct {
	State EKeyState `xml:",attr"`
}

type Port struct {
	Id       int      `xml:",attr"`
	Address  string   `xml:",attr"`
	Type     PortType `xml:",attr"`
	Upstream bool     `xml:",attr"`
	Bus      struct {
		Size int `xml:",attr,omitempty"`
	}
}

type Communications struct {
	ConfigTag   ConfigTag
	Connections []Connection `xml:"Connections>Connection"`
}

type ConfigTag struct {
	ConfigSize     int            `xml:",attr"`
	ExternalAccess ExternalAccess `xml:",attr"`
	Data           []Data
}

type Data struct {
	Format    DataFormat      `xml:",attr"`
	L5K       string          `xml:",cdata"`
	Structure Structure       `xml:",omitempty"`
	DataValue DataValueMember `xml:",omitempty"`
	Array     Array           `xml:",omitempty"`
}

type Structure struct {
	DataType        string            `xml:",attr"` // TODO: enum
	StructureMember []DataValueMember `xml:"StructureMember>DataValueMember"`
	DataValueMember []DataValueMember
	ArrayMember     Array `xml:",omitempty"`
}

type DataValueMember struct {
	Name     string `xml:",attr,omitempty"`
	DataType string `xml:",attr"` // TODO: enum
	Radix    Radix  `xml:",attr"`
	Value    string `xml:",attr"`
}

type Array struct {
	Name       string    `xml:",attr,omitempty"`
	DataType   string    `xml:",attr"` // TODO: enum
	Dimensions ArrayDims `xml:",attr"`
	Radix      Radix     `xml:",attr"`
	Elements   []Element `xml:"Element"`
}

type Element struct {
	Index Index  `xml:",attr"`
	Value string `xml:",attr"`
}

type Connection struct {
	Name        string `xml:",attr"`
	RPI         int    `xml:",attr"`
	Type        IOType `xml:",attr"`
	EventID     int    `xml:",attr"`
	SendTrigger bool   `xml:"ProgrammaticallySendEventTrigger,attr"`
	InputTag    IOTag
	OutputTag   IOTag
}

type IOTag struct {
	ExternalAccess ExternalAccess `xml:",attr"`
	Comments       []Comment      `xml:"Comments>Comment,omitempty"`
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

type AddOnInstrDef struct {
	Name                 string      `xml:",attr"`
	Revision             string      `xml:",attr"`
	RevisionExtension    string      `xml:",attr"`
	Vendor               string      `xml:",attr"`
	ExecutePrescan       bool        `xml:",attr"`
	ExecutePostscan      bool        `xml:",attr"`
	ExecuteEnableInFalse bool        `xml:",attr"`
	CreatedDate          iso8601Time `xml:",attr"`
	CreatedBy            string      `xml:",attr"`
	EditedDate           iso8601Time `xml:",attr"`
	EditedBy             string      `xml:",attr"`
	SoftwareRevision     string      `xml:",attr"`
	Description          Description `xml:",omitempty"`
	Parameters           []Parameter `xml:"Parameters>Parameter"`
	LocalTags            []LocalTag  `xml:"LocalTags>LocalTag"`
	Routines             []Routine   `xml:"Routines>Routine"`
}

type Parameter struct {
	Name           string         `xml:",attr"`
	TagType        TagType        `xml:",attr"`
	DataType       string         `xml:",attr"` // TODO: enum
	Usage          IOType         `xml:",attr"`
	Radix          Radix          `xml:",attr"`
	Required       bool           `xml:",attr"`
	Visible        bool           `xml:",attr"`
	ExternalAccess ExternalAccess `xml:",attr"`
	Description    Description    `xml:",omitempty"`
	DefaultData    []Data         `xml:",omitempty"`
}

type LocalTag struct {
	Name           string         `xml:",attr"`
	DataType       string         `xml:",attr"` // TODO: enum
	Dimensions     int            `xml:",attr"`
	Radix          Radix          `xml:",attr"`
	ExternalAccess ExternalAccess `xml:",attr"`
	Description    Description
	DefaultData    []Data `xml:",omitempty"`
}

type Routine struct {
	Name        string      `xml:",attr"`
	Type        RoutineType `xml:",attr"`
	Description Description
	RLLContent  struct {
		Rungs []Rung `xml:"Rung"`
	}
}

type Rung struct {
	Number  int      `xml:",attr"`
	Type    RungType `xml:",attr"`
	Comment Description
	Text    Description
}

type TagList []Tag

type Tag struct {
	Name           string         `xml:",attr"`
	TagType        TagType        `xml:",attr"`
	DataType       string         `xml:",attr"` // TODO: enum
	Dimensions     TagDims        `xml:",attr,omitempty"`
	Radix          Radix          `xml:",attr,omitempty"`
	Constant       bool           `xml:",attr"`
	ExternalAccess ExternalAccess `xml:",attr"`
	Description    Description
	Data           []Data
}

type Program struct {
	Name            string    `xml:",attr"`
	TestEdits       bool      `xml:",attr"`
	MainRoutineName string    `xml:",attr"`
	Disabled        bool      `xml:",attr"`
	UseAsFolder     bool      `xml:",attr"`
	Tags            TagList   `xml:"Tags>Tag"`
	Routines        []Routine `xml:"Routines>Routine"`
}

type Task struct {
	Name                 string   `xml:",attr"`
	Type                 TaskType `xml:",attr"`
	Priority             int      `xml:",attr"`
	Watchdog             int      `xml:",attr"`
	DisableUpdateOutputs bool     `xml:",attr"`
	InhibitTask          bool     `xml:",attr"`
	ScheduledPrograms    []struct {
		Name string `xml:",attr"`
	} `xml:"ScheduledPrograms>ScheduledProgram"`
}

type CST struct {
	MasterID int `xml:",attr"`
}

type WallClockTime struct {
	LocalTimeAdjustment int `xml:",attr"`
	TimeZone            int `xml:",attr"`
}

type Trend struct {
	Name             string `xml:",attr"`
	SamplePeriod     int    `xml:",attr"`
	NumberOfCaptures int    `xml:",attr"`
	CaptureSizeType  string `xml:",attr"`
	CaptureSize      int    `xml:",attr"`
	StartTriggerType string `xml:",attr"`
	StopTriggerType  string `xml:",attr"`
	TrendxVersion    string `xml:",attr"`
	Template         string
	Pens             []Pen `xml:"Pens>Pen"`
}

type Pen struct {
	Name    string  `xml:",attr"`
	Color   string  `xml:",attr"`
	Visible bool    `xml:",attr"`
	Style   int     `xml:",attr"`
	Type    string  `xml:",attr"`
	Width   int     `xml:",attr"`
	Marker  int     `xml:",attr"`
	Min     float32 `xml:",attr"`
	Max     float32 `xml:",attr"`
}

type TimeSynchronize struct {
	Priority1 int  `xml:",attr"`
	Priority2 int  `xml:",attr"`
	PTPEnable bool `xml:",attr"`
}

type EthernetPort struct {
	Port        int  `xml:",attr"`
	Label       int  `xml:",attr"`
	PortEnabled bool `xml:",attr"`
}

type EthernetNetwork struct {
	SupervisorModeEnabled bool `xml:",attr"`
	SupervisorPrecedence  int  `xml:",attr"`
	BeaconInterval        int  `xml:",attr"`
	BeaconTimeout         int  `xml:",attr"`
	VLANID                int  `xml:",attr"`
}

const (
	iso8601Format = "2006-01-02T15:04:05.000Z"
	rsLogixFormat = "Mon Jan 2 15:04:05 2006"
)

type rsLogixTime time.Time

func (rlt *rsLogixTime) fromString(str string) error {
	parse, err := time.Parse(rsLogixFormat, str)
	if err != nil {
		return err
	}
	*rlt = rsLogixTime(parse)
	return nil
}
func (rlt rsLogixTime) toString() string {
	return time.Time(rlt).Format(rsLogixFormat)
}
func (rlt *rsLogixTime) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	d.DecodeElement(&str, &start)
	return rlt.fromString(str)
}
func (rlt rsLogixTime) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(rlt.toString(), start)
}
func (rlt *rsLogixTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return rlt.fromString(attr.Value)
}
func (rlt rsLogixTime) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: rlt.toString()}, nil
}

type iso8601Time time.Time

func (rlt *iso8601Time) fromString(str string) error {
	parse, err := time.Parse(iso8601Format, str)
	if err != nil {
		return err
	}
	*rlt = iso8601Time(parse)
	return nil
}
func (rlt iso8601Time) toString() string {
	return time.Time(rlt).Format(iso8601Format)
}
func (rlt *iso8601Time) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	return rlt.fromString(str)
}
func (rlt iso8601Time) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(rlt.toString(), start)
}
func (rlt *iso8601Time) UnmarshalXMLAttr(attr xml.Attr) error {
	return rlt.fromString(attr.Value)
}
func (rlt iso8601Time) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: rlt.toString()}, nil
}

// stringSlice is imported from a string of space-separated strings
type stringSlice struct{ strings []string }

func (ss *stringSlice) fromString(str string) error {
	for _, st := range strings.Split(str, " ") {
		ss.strings = append(ss.strings, st)
	}
	return nil
}
func (ss stringSlice) toString() string {
	return strings.Join(ss.strings, " ")
}
func (ss *stringSlice) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	return ss.fromString(str)
}
func (ss stringSlice) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(ss.toString(), start)
}
func (ss *stringSlice) UnmarshalXMLAttr(attr xml.Attr) error {
	return ss.fromString(attr.Value)
}
func (ss stringSlice) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: ss.toString()}, nil
}

type TagDims []int
type ArrayDims TagDims

func (dims *TagDims) fromString(str string, sep string) error {
	strs := strings.Split(str, sep)
	ints := make([]int, len(strs))
	for i, str := range strs {
		val, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			return err
		}
		ints[i] = int(val)
	}
	*dims = ints
	return nil
}
func (dims TagDims) toString(sep string) string {
	strs := make([]string, len(dims))
	for i, val := range dims {
		strs[i] = strconv.Itoa(val)
	}
	return strings.Join(strs, sep)
}
func (dims *TagDims) UnmarshalXMLAttr(attr xml.Attr) error {
	return dims.fromString(attr.Value, " ")
}
func (dims TagDims) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: dims.toString(" ")}, nil
}
func (dims *ArrayDims) UnmarshalXMLAttr(attr xml.Attr) error {
	return (*TagDims)(dims).fromString(attr.Value, ",")
}
func (dims ArrayDims) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: TagDims(dims).toString(",")}, nil
}

type Index TagDims

func (dims *Index) UnmarshalXMLAttr(attr xml.Attr) error {
	if attr.Value[0] != '[' || attr.Value[len(attr.Value)-1] != ']' {
		return fmt.Errorf("index attribute '%s' should be enclosed by brackets", attr.Value)
	}
	return (*TagDims)(dims).fromString(attr.Value[1:len(attr.Value)-1], ",")
}
func (dims Index) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: "[" + TagDims(dims).toString(" ") + "]"}, nil
}

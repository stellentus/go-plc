package l5x

import (
	"encoding/xml"
)

type ExternalAccess int

const (
	ExternalAccessReadWrite ExternalAccess = iota
	ExternalAccessReadOnly
)

var externalAccessNames = []string{"Read/Write", "Read Only"}

func (enum *ExternalAccess) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), externalAccessNames, d, start)
}
func (enum ExternalAccess) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), externalAccessNames, e, start)
}
func (enum *ExternalAccess) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), externalAccessNames, attr)
}
func (enum ExternalAccess) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), externalAccessNames, name)
}
func (enum ExternalAccess) String() string { return enumToString(int(enum), externalAccessNames) }

type Class int

const (
	ClassUser Class = iota
)

var classNames = []string{"User"}

func (enum *Class) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), classNames, d, start)
}
func (enum Class) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), classNames, e, start)
}
func (enum *Class) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), classNames, attr)
}
func (enum Class) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), classNames, name)
}
func (enum Class) String() string { return enumToString(int(enum), classNames) }

type Radix int

const (
	RadixDecimal Radix = iota
	RadixNullType
	RadixFloat
	RadixBinary
	RadixASCII
	RadixHex
)

var radixNames = []string{"Decimal", "NullType", "Float", "Binary", "ASCII", "Hex"}

func (enum *Radix) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), radixNames, d, start)
}
func (enum Radix) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), radixNames, e, start)
}
func (enum *Radix) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), radixNames, attr)
}
func (enum Radix) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), radixNames, name)
}
func (enum Radix) String() string { return enumToString(int(enum), radixNames) }

type PortType int

const (
	PortTypeEthernet PortType = iota
	PortTypeCompact
)

var portTypeNames = []string{"Ethernet", "Compact"}

func (enum *PortType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), portTypeNames, d, start)
}
func (enum PortType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), portTypeNames, e, start)
}
func (enum *PortType) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), portTypeNames, attr)
}
func (enum PortType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), portTypeNames, name)
}
func (enum PortType) String() string { return enumToString(int(enum), portTypeNames) }

type DataFormat int

const (
	DataFormatDecorated DataFormat = iota
	DataFormatL5K
	DataFormatMessage
)

var dataFormatNames = []string{"Decorated", "L5K", "Message"}

func (enum *DataFormat) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), dataFormatNames, d, start)
}
func (enum DataFormat) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), dataFormatNames, e, start)
}
func (enum *DataFormat) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), dataFormatNames, attr)
}
func (enum DataFormat) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), dataFormatNames, name)
}
func (enum DataFormat) String() string { return enumToString(int(enum), dataFormatNames) }

type TaskType int

const (
	TaskTypeContinuous TaskType = iota
)

var taskTypeNames = []string{"CONTINUOUS"}

func (enum *TaskType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), taskTypeNames, d, start)
}
func (enum TaskType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), taskTypeNames, e, start)
}
func (enum *TaskType) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), taskTypeNames, attr)
}
func (enum TaskType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), taskTypeNames, name)
}
func (enum TaskType) String() string { return enumToString(int(enum), taskTypeNames) }

type DataTypeFamily int

const (
	DataTypeFamilyNone DataTypeFamily = iota
	DataTypeFamilyString
)

var dataTypeFamilyNames = []string{"NoFamily", "StringFamily"}

func (enum *DataTypeFamily) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), dataTypeFamilyNames, d, start)
}
func (enum DataTypeFamily) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), dataTypeFamilyNames, e, start)
}
func (enum *DataTypeFamily) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), dataTypeFamilyNames, attr)
}
func (enum DataTypeFamily) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), dataTypeFamilyNames, name)
}
func (enum DataTypeFamily) String() string { return enumToString(int(enum), dataTypeFamilyNames) }

type EKeyState int

const (
	EKeyStateExactMatch EKeyState = iota
	EKeyStateCompatibleModule
	EKeyStateDisabled
)

var eKeyStateNames = []string{"ExactMatch", "CompatibleModule", "Disabled"}

func (enum *EKeyState) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), eKeyStateNames, d, start)
}
func (enum EKeyState) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), eKeyStateNames, e, start)
}
func (enum *EKeyState) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), eKeyStateNames, attr)
}
func (enum EKeyState) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), eKeyStateNames, name)
}
func (enum EKeyState) String() string { return enumToString(int(enum), eKeyStateNames) }

type IOType int

const (
	IOTypeInput IOType = iota
	IOTypeOutput
)

var ioTypeNames = []string{"Input", "Output"}

func (enum *IOType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), ioTypeNames, d, start)
}
func (enum IOType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), ioTypeNames, e, start)
}
func (enum *IOType) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), ioTypeNames, attr)
}
func (enum IOType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), ioTypeNames, name)
}
func (enum IOType) String() string { return enumToString(int(enum), ioTypeNames) }

type TagType int

const (
	TagTypeBase TagType = iota
)

var tagTypeNames = []string{"Base"}

func (enum *TagType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), tagTypeNames, d, start)
}
func (enum TagType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), tagTypeNames, e, start)
}
func (enum *TagType) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), tagTypeNames, attr)
}
func (enum TagType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), tagTypeNames, name)
}
func (enum TagType) String() string { return enumToString(int(enum), tagTypeNames) }

type RoutineType int

const (
	RoutineTypeRLL RoutineType = iota
	RoutineTypeST
	RoutineTypeFBD
)

var routineTypeNames = []string{"RLL", "ST", "FBD"}

func (enum *RoutineType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), routineTypeNames, d, start)
}
func (enum RoutineType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), routineTypeNames, e, start)
}
func (enum *RoutineType) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), routineTypeNames, attr)
}
func (enum RoutineType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), routineTypeNames, name)
}
func (enum RoutineType) String() string { return enumToString(int(enum), routineTypeNames) }

type RungType int

const (
	RungTypeN RungType = iota
)

var rungTypeNames = []string{"N"}

func (enum *RungType) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	return enumUnmarshalXML((*int)(enum), rungTypeNames, d, start)
}
func (enum RungType) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	return enumMarshalXML(int(enum), rungTypeNames, e, start)
}
func (enum *RungType) UnmarshalXMLAttr(attr xml.Attr) error {
	return enumUnmarshalXMLAttr((*int)(enum), rungTypeNames, attr)
}
func (enum RungType) MarshalXMLAttr(name xml.Name) (xml.Attr, error) {
	return enumMarshalXMLAttr(int(enum), rungTypeNames, name)
}
func (enum RungType) String() string { return enumToString(int(enum), rungTypeNames) }

// The following functions are designed to make it easy to marshal/unmarshal enums.
// It always uses -1 and "Unknown" to indicate an unknown value. Otherwise,
// the enum is assumed to start at 0 without skipping any values.
// This would probably all be better using go generate.
func enumFromString(enum *int, names []string, str string) error {
	for i, estr := range names {
		if str == estr {
			*enum = i
			return nil
		}
	}
	*enum = -1
	return nil
}
func enumToString(enum int, names []string) string {
	if enum < 0 || enum >= len(names) {
		return "Unknown"
	}
	return names[enum]
}
func enumUnmarshalXML(enum *int, names []string, d *xml.Decoder, start xml.StartElement) error {
	var str string
	err := d.DecodeElement(&str, &start)
	if err != nil {
		return err
	}
	return enumFromString(enum, names, str)
}
func enumMarshalXML(enum int, names []string, e *xml.Encoder, start xml.StartElement) error {
	return e.EncodeElement(enumToString(enum, names), start)
}
func enumUnmarshalXMLAttr(enum *int, names []string, attr xml.Attr) error {
	return enumFromString(enum, names, attr.Value)
}
func enumMarshalXMLAttr(enum int, names []string, name xml.Name) (xml.Attr, error) {
	return xml.Attr{Name: name, Value: enumToString(enum, names)}, nil
}

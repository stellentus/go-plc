package l5x

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Type interface {
	// PlcName is the name of the type, as used by the PLC (e.g. "my_struct" or "INT")
	PlcName() string

	// GoName is the name of the type, as used by go (e.g. "MyStruct" or "int16")
	GoName() string

	// GoTypeString is the go definition of the type (e.g. "struct {Val uint8}" or "").
	// For basic types, it's an empty string.
	GoTypeString() string
}

type NamedType struct {
	PlcName string
	GoName  string
	Type    Type
}

type TypeList []Type

func NamedTypeDeclaration(nt NamedType) string {
	return nt.GoName + " " + nt.Type.GoName()
}

func TypeDefinition(ty Type) string {
	if ty.GoTypeString() == "" {
		// This type shouldn't be defined
		return ""
	}
	return "type " + makeValidIdentifier(ty.GoName()) + " " + ty.GoTypeString()
}

var (
	typeBOOL   = basicType{"BOOL", "bool"}
	typeSINT   = basicType{"SINT", "int8"}
	typeINT    = basicType{"INT", "int16"}
	typeDINT   = basicType{"DINT", "int32"}
	typeLINT   = basicType{"LINT", "int64"}
	typeUSINT  = basicType{"USINT", "uint8"}
	typeUINT   = basicType{"UINT", "uint16"}
	typeUDINT  = basicType{"UDINT", "uint32"}
	typeULINT  = basicType{"ULINT", "uint64"}
	typeREAL   = basicType{"REAL", "float32"}
	typeLREAL  = basicType{"LREAL", "float64"}
	typeSTRING = basicType{"STRING", "string"}
	typeBYTE   = basicType{"BYTE", "byte"}    // might not be correct; might be an array
	typeWORD   = basicType{"WORD", "uint16"}  // might not be correct; might be an array
	typeDWORD  = basicType{"DWORD", "uint32"} // might not be correct; might be an array
	typeLWORD  = basicType{"LWORD", "uint64"} // might not be correct; might be an array
)

func NewTypeList() TypeList {
	return TypeList{
		typeBOOL,
		typeSINT,
		typeINT,
		typeDINT,
		typeLINT,
		typeUSINT,
		typeUINT,
		typeUDINT,
		typeULINT,
		typeREAL,
		typeLREAL,
		typeSTRING,
		typeBYTE,
		typeWORD,
		typeDWORD,
		typeLWORD,
	}
}

func (tl TypeList) WithPlcName(name string) (Type, error) {
	for _, ty := range tl {
		if ty.PlcName() == name {
			return ty, nil
		}
	}
	return nil, errors.New("DataType '" + name + "' couldn't be found")
}

func (tl TypeList) WriteDefinitions(wr io.Writer) error {
	for _, ty := range tl {
		str := TypeDefinition(ty)
		if str == "" {
			continue
		}
		_, err := wr.Write([]byte(str + "\n\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

// AddControlLogixTypes adds a couple of types that are (I think) built into ControlLogix
func (tl *TypeList) AddControlLogixTypes() error {
	timer, err := DataType{
		Name: "TIMER",
		Members: []Member{
			{Name: "PRE", DataType: "DINT"},
			{Name: "ACC", DataType: "DINT"},
			{Name: "EN", DataType: "BOOL"},
			{Name: "TT", DataType: "BOOL"},
			{Name: "DN", DataType: "BOOL"},
		},
	}.AsType(*tl)
	if err != nil {
		return fmt.Errorf("could not register ControlLogix TIMER because %w", err)
	}
	*tl = append(*tl, timer)

	counter, err := DataType{
		Name: "COUNTER",
		Members: []Member{
			{Name: "PRE", DataType: "DINT"},
			{Name: "ACC", DataType: "DINT"},
			{Name: "CU", DataType: "BOOL"},
			{Name: "CD", DataType: "BOOL"},
			{Name: "DN", DataType: "BOOL"},
			{Name: "OV", DataType: "BOOL"},
			{Name: "UN", DataType: "BOOL"},
		},
	}.AsType(*tl)
	if err != nil {
		return fmt.Errorf("could not register ControlLogix COUNTER because %w", err)
	}
	*tl = append(*tl, counter)

	return nil
}

func (mb Member) AsNamedType(knownTypes TypeList) (NamedType, error) {
	return knownTypes.newNamedType(mb.Name, mb.DataType, []int{mb.Dimension})
}

func (tl TypeList) newNamedType(name, dataType string, dims []int) (NamedType, error) {
	var nt NamedType
	for _, ty := range tl {
		if ty.PlcName() == dataType {
			nt = NamedType{
				GoName: name,
				Type:   ty,
			}
			if !isPublicGoIdentifier(name) {
				valid := makeValidIdentifier(nt.GoName)
				if valid == "" {
					return NamedType{}, fmt.Errorf("couldn't create valid identifier for '%s'", name)
				}
				nt.PlcName = nt.GoName
				nt.GoName = valid
			}
			break
		}
	}
	if nt.Type == nil {
		return NamedType{}, ErrUnknownType
	}

	if len(dims) == 1 && dims[0] <= 1 { // not an array
		return nt, nil
	}

	for idx := len(dims) - 1; idx >= 0; idx-- {
		if dims[idx] <= 1 {
			return NamedType{}, fmt.Errorf("couldn't create NamedType with dimensions %v", dims)
		}
		nt.Type = arrayType{
			elementInfo: nt.Type,
			count:       dims[idx],
		}
	}

	return nt, nil
}

func (dt DataType) AsType(knownTypes TypeList) (Type, error) {
	if dt.Class != ClassUser {
		return nil, fmt.Errorf("Unknown class type")
	}

	switch dt.Family {
	case DataTypeFamilyString:
		return parseString(dt.Name, dt.Members)

	case DataTypeFamilyNone:
		return parseStruct(dt.Name, dt.Members, knownTypes)

	default:
		return nil, fmt.Errorf("Unknown data family type")
	}
}

func parseStruct(name string, membs []Member, knownTypes TypeList) (Type, error) {
	sti := structType{
		name:    name,
		members: []NamedType{},
	}
	for _, memb := range membs {
		if memb.DataType == "BIT" {
			// Not yet implemented, so ignore it
			continue
		}
		nm, err := memb.AsNamedType(knownTypes)
		if err != nil {
			if errors.Is(err, ErrUnknownType) {
				err = errUnknownTypeSpecific{name, memb.DataType}
			}
			return nil, err
		}
		sti.members = append(sti.members, nm)
	}
	if len(sti.members) == 0 {
		return nil, fmt.Errorf("DataType '%s' produced no members (%d)", name, len(membs))
	}

	return sti, nil
}

func parseString(name string, memb []Member) (Type, error) {
	if len(memb) != 2 {
		return nil, fmt.Errorf("StringFamily '%s' had %d members instead of 2", name, len(memb))
	}

	if memb[0].Name != "LEN" {
		return nil, fmt.Errorf("StringFamily '%s' LEN member is missing: %s", name, memb[0].Name)
	}
	if memb[0].DataType != "DINT" {
		return nil, fmt.Errorf("StringFamily '%s' LEN.DataType is incorrect: %v", name, memb[0].DataType)
	}
	if memb[0].Dimension != 0 {
		return nil, fmt.Errorf("StringFamily '%s' LEN.Dimension is incorrect: %d", name, memb[0].Dimension)
	}
	if memb[0].Radix != RadixDecimal {
		return nil, fmt.Errorf("StringFamily '%s' LEN.Radix is incorrect: %v", name, memb[0].Radix)
	}

	if memb[1].Name != "DATA" {
		return nil, fmt.Errorf("StringFamily '%s' DATA member is missing: %v", name, memb[1].Name)
	}
	if memb[1].DataType != "SINT" {
		return nil, fmt.Errorf("StringFamily '%s' DATA.DataType is incorrect: %v", name, memb[1].DataType)
	}
	if memb[1].Dimension <= 0 {
		return nil, fmt.Errorf("StringFamily '%s' DATA.Dimension is invalid: %d", name, memb[1].Dimension)
	}
	if memb[1].Radix != RadixASCII {
		return nil, fmt.Errorf("StringFamily '%s' DATA.Radix is incorrect: %v", name, memb[1].Radix)
	}

	return stringType{
		name: name,
		atype: arrayType{
			elementInfo: basicType{"SINT", "int8"},
			count:       memb[1].Dimension,
		},
	}, nil
}

type basicType struct {
	plcName, goName string
}

func (bt basicType) PlcName() string      { return bt.plcName }
func (bt basicType) GoName() string       { return bt.goName }
func (bt basicType) GoTypeString() string { return "" }

type arrayType struct {
	elementInfo Type
	count       int
}

func (ati arrayType) PlcName() string { return "" }
func (ati arrayType) GoName() string  { return ati.GoTypeString() }
func (ati arrayType) GoTypeString() string {
	return fmt.Sprintf("[%d]%v", ati.count, ati.elementInfo.GoName())
}

type structType struct {
	name    string
	members []NamedType
}

func (sti structType) PlcName() string { return sti.name }
func (sti structType) GoName() string  { return sti.name }
func (sti structType) GoTypeString() string {
	strs := make([]string, len(sti.members))
	for i, member := range sti.members {
		tagSuffix := ""
		if member.PlcName != "" {
			tagSuffix = " `plc:\"" + member.PlcName + "\"`"
		}
		strs[i] = fmt.Sprintf("\n\t%s%s", NamedTypeDeclaration(member), tagSuffix)
	}
	return fmt.Sprintf("struct {%s\n}", strings.Join(strs, ""))
}

type stringType struct {
	name  string
	atype arrayType
}

func (sty stringType) PlcName() string      { return sty.name }
func (sty stringType) GoName() string       { return sty.name }
func (sty stringType) GoTypeString() string { return sty.atype.GoTypeString() }

// isValidGoIdentifier determines if str is a valid go identifier. According to the spec:
//     identifier = letter { letter | unicode_digit } .
//     letter        = unicode_letter | "_" .
//     unicode_letter = /* a Unicode code point classified as "Letter" */ .
//     unicode_digit  = /* a Unicode code point classified as "Number, decimal digit" */ .
func isValidGoIdentifier(str string) bool {
	if len(str) <= 0 {
		return false
	}
	isLetter := func(r rune) bool {
		if unicode.IsLetter(r) {
			return true
		}
		return r == '_'
	}
	for pos, r := range str {
		if isLetter(r) {
			continue // A letter is always allowed
		}
		if unicode.IsDigit(r) && pos > 0 {
			continue // After the first position, a digit is allowed
		}
		return false
	}
	return true
}

// isPublicGoIdentifier checks if str is public (i.e. starts with a capital)
// and a valid go identifier.
func isPublicGoIdentifier(str string) bool {
	if !isValidGoIdentifier(str) {
		return false
	}
	for _, r := range str {
		return unicode.IsUpper(r)
	}
	return false
}

// makeValidIdentifier returns a valid go identifer from 'str'.
// It returns "" if it can't make an identifier.
func makeValidIdentifier(str string) string {
	var valid strings.Builder
	for pos, r := range strings.TrimSpace(str) {
		if pos == 0 {
			valid.WriteRune(unicode.ToUpper(r))
			continue
		}
		if !isValidIdentiferRune(r) {
			valid.WriteRune('_')
			continue
		}
		valid.WriteRune(r)
	}
	if !isValidGoIdentifier(valid.String()) {
		return ""
	}
	return valid.String()
}

// isValidIdentiferRune checks if runes after the first one are valid.
func isValidIdentiferRune(r rune) bool {
	if unicode.IsLetter(r) {
		return true
	}
	if unicode.IsDigit(r) {
		return true
	}
	return r == '_'
}

var ErrUnknownType = errors.New("unknown type")

type errUnknownTypeSpecific struct {
	typ, requires string
}

func (err errUnknownTypeSpecific) Error() string {
	return fmt.Sprintf("Type '%s' requires type '%s'", err.typ, err.requires)
}

func (err errUnknownTypeSpecific) Unwrap() error {
	return ErrUnknownType
}

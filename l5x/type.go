package l5x

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/stellentus/go-plc"
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

// NamedTypeDeclaration is used to declare a named type inside a struct.
// It automatically embeds if the variable name matches the type name.
func NamedTypeDeclaration(nt NamedType) string {
	typeName := nt.Type.GoName()
	if nt.GoName == typeName {
		return nt.GoName
	}
	return nt.GoName + " " + typeName
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

	pidEnhanced, err := DataType{
		Name: "PID_ENHANCED",
		Members: []Member{
			{Name: "EnableIn", DataType: "BOOL"},
			{Name: "PV", DataType: "REAL"},
			{Name: "PVFault", DataType: "BOOL"},
			{Name: "PVEUMax", DataType: "REAL"},
			{Name: "PVEUMin", DataType: "REAL"},
			{Name: "SPProg", DataType: "REAL"},
			{Name: "SPOper", DataType: "REAL"},
			{Name: "SPCascade", DataType: "REAL"},
			{Name: "SPHLimit", DataType: "REAL"},
			{Name: "SPLLimit", DataType: "REAL"},
			{Name: "UseRatio", DataType: "BOOL"},
			{Name: "RatioProg", DataType: "REAL"},
			{Name: "RatioOper", DataType: "REAL"},
			{Name: "RatioHLimit", DataType: "REAL"},
			{Name: "RatioLLimit", DataType: "REAL"},
			{Name: "CVFault", DataType: "BOOL"},
			{Name: "CVInitReq", DataType: "BOOL"},
			{Name: "CVInitValue", DataType: "REAL"},
			{Name: "CVProg", DataType: "REAL"},
			{Name: "CVOper", DataType: "REAL"},
			{Name: "CVOverride", DataType: "REAL"},
			{Name: "CVPrevious", DataType: "REAL"},
			{Name: "CVSetPrevious", DataType: "BOOL"},
			{Name: "CVManLimiting", DataType: "BOOL"},
			{Name: "CVEUMax", DataType: "REAL"},
			{Name: "CVEUMin", DataType: "REAL"},
			{Name: "CVHLimit", DataType: "REAL"},
			{Name: "CVLLimit", DataType: "REAL"},
			{Name: "CVROCLimit", DataType: "REAL"},
			{Name: "FF", DataType: "REAL"},
			{Name: "FFPrevious", DataType: "REAL"},
			{Name: "FFSetPrevious", DataType: "BOOL"},
			{Name: "HandFB", DataType: "REAL"},
			{Name: "HandFBFault", DataType: "BOOL"},
			{Name: "WindupHIn", DataType: "BOOL"},
			{Name: "WindupLIn", DataType: "BOOL"},
			{Name: "ControlAction", DataType: "BOOL"},
			{Name: "DependIndepend", DataType: "BOOL"},
			{Name: "PGain", DataType: "REAL"},
			{Name: "IGain", DataType: "REAL"},
			{Name: "DGain", DataType: "REAL"},
			{Name: "PVEProportional", DataType: "BOOL"},
			{Name: "PVEDerivative", DataType: "BOOL"},
			{Name: "DSmoothing", DataType: "BOOL"},
			{Name: "PVTracking", DataType: "BOOL"},
			{Name: "ZCDeadband", DataType: "REAL"},
			{Name: "ZCOff", DataType: "BOOL"},
			{Name: "PVHHLimit", DataType: "REAL"},
			{Name: "PVHLimit", DataType: "REAL"},
			{Name: "PVLLimit", DataType: "REAL"},
			{Name: "PVLLLimit", DataType: "REAL"},
			{Name: "PVDeadband", DataType: "REAL"},
			{Name: "PVROCPosLimit", DataType: "REAL"},
			{Name: "PVROCNegLimit", DataType: "REAL"},
			{Name: "PVROCPeriod", DataType: "REAL"},
			{Name: "DevHHLimit", DataType: "REAL"},
			{Name: "DevHLimit", DataType: "REAL"},
			{Name: "DevLLimit", DataType: "REAL"},
			{Name: "DevLLLimit", DataType: "REAL"},
			{Name: "DevDeadband", DataType: "REAL"},
			{Name: "AllowCasRat", DataType: "BOOL"},
			{Name: "ManualAfterInit", DataType: "BOOL"},
			{Name: "ProgProgReq", DataType: "BOOL"},
			{Name: "ProgOperReq", DataType: "BOOL"},
			{Name: "ProgCasRatReq", DataType: "BOOL"},
			{Name: "ProgAutoReq", DataType: "BOOL"},
			{Name: "ProgManualReq", DataType: "BOOL"},
			{Name: "ProgOverrideReq", DataType: "BOOL"},
			{Name: "ProgHandReq", DataType: "BOOL"},
			{Name: "OperProgReq", DataType: "BOOL"},
			{Name: "OperOperReq", DataType: "BOOL"},
			{Name: "OperCasRatReq", DataType: "BOOL"},
			{Name: "OperAutoReq", DataType: "BOOL"},
			{Name: "OperManualReq", DataType: "BOOL"},
			{Name: "ProgValueReset", DataType: "BOOL"},
			{Name: "TimingMode", DataType: "DINT"},
			{Name: "OversampleDT", DataType: "REAL"},
			{Name: "RTSTime", DataType: "DINT"},
			{Name: "RTSTimeStamp", DataType: "DINT"},
			{Name: "AtuneAcquire", DataType: "BOOL"},
			{Name: "AtuneStart", DataType: "BOOL"},
			{Name: "AtuneUseGains", DataType: "BOOL"},
			{Name: "AtuneAbort", DataType: "BOOL"},
			{Name: "AtuneUnacquire", DataType: "BOOL"},
			{Name: "EnableOut", DataType: "BOOL"},
			{Name: "CVEU", DataType: "REAL"},
			{Name: "CV", DataType: "REAL"},
			{Name: "CVInitializing", DataType: "BOOL"},
			{Name: "CVHAlarm", DataType: "BOOL"},
			{Name: "CVLAlarm", DataType: "BOOL"},
			{Name: "CVROCAlarm", DataType: "BOOL"},
			{Name: "SP", DataType: "REAL"},
			{Name: "SPPercent", DataType: "REAL"},
			{Name: "SPHAlarm", DataType: "BOOL"},
			{Name: "SPLAlarm", DataType: "BOOL"},
			{Name: "PVPercent", DataType: "REAL"},
			{Name: "E", DataType: "REAL"},
			{Name: "EPercent", DataType: "REAL"},
			{Name: "InitPrimary", DataType: "BOOL"},
			{Name: "WindupHOut", DataType: "BOOL"},
			{Name: "WindupLOut", DataType: "BOOL"},
			{Name: "Ratio", DataType: "REAL"},
			{Name: "RatioHAlarm", DataType: "BOOL"},
			{Name: "RatioLAlarm", DataType: "BOOL"},
			{Name: "ZCDeadbandOn", DataType: "BOOL"},
			{Name: "PVHHAlarm", DataType: "BOOL"},
			{Name: "PVHAlarm", DataType: "BOOL"},
			{Name: "PVLAlarm", DataType: "BOOL"},
			{Name: "PVLLAlarm", DataType: "BOOL"},
			{Name: "PVROCPosAlarm", DataType: "BOOL"},
			{Name: "PVROCNegAlarm", DataType: "BOOL"},
			{Name: "DevHHAlarm", DataType: "BOOL"},
			{Name: "DevHAlarm", DataType: "BOOL"},
			{Name: "DevLAlarm", DataType: "BOOL"},
			{Name: "DevLLAlarm", DataType: "BOOL"},
			{Name: "ProgOper", DataType: "BOOL"},
			{Name: "CasRat", DataType: "BOOL"},
			{Name: "Auto", DataType: "BOOL"},
			{Name: "Manual", DataType: "BOOL"},
			{Name: "Override", DataType: "BOOL"},
			{Name: "Hand", DataType: "BOOL"},
			{Name: "DeltaT", DataType: "REAL"},
			{Name: "AtuneReady", DataType: "BOOL"},
			{Name: "AtuneOn", DataType: "BOOL"},
			{Name: "AtuneDone", DataType: "BOOL"},
			{Name: "AtuneAborted", DataType: "BOOL"},
			{Name: "AtuneBusy", DataType: "BOOL"},
			{Name: "Status1", DataType: "DINT"},
			{Name: "Status2", DataType: "DINT"},
			{Name: "InstructFault", DataType: "BOOL"},
			{Name: "PVFaulted", DataType: "BOOL"},
			{Name: "CVFaulted", DataType: "BOOL"},
			{Name: "HandFBFaulted", DataType: "BOOL"},
			{Name: "PVSpanInv", DataType: "BOOL"},
			{Name: "SPProgInv", DataType: "BOOL"},
			{Name: "SPOperInv", DataType: "BOOL"},
			{Name: "SPCascadeInv", DataType: "BOOL"},
			{Name: "SPLimitsInv", DataType: "BOOL"},
			{Name: "RatioProgInv", DataType: "BOOL"},
			{Name: "RatioOperInv", DataType: "BOOL"},
			{Name: "RatioLimitsInv", DataType: "BOOL"},
			{Name: "CVProgInv", DataType: "BOOL"},
			{Name: "CVOperInv", DataType: "BOOL"},
			{Name: "CVOverrideInv", DataType: "BOOL"},
			{Name: "CVPreviousInv", DataType: "BOOL"},
			{Name: "CVEUSpanInv", DataType: "BOOL"},
			{Name: "CVLimitsInv", DataType: "BOOL"},
			{Name: "CVROCLimitInv", DataType: "BOOL"},
			{Name: "FFInv", DataType: "BOOL"},
			{Name: "FFPreviousInv", DataType: "BOOL"},
			{Name: "HandFBInv", DataType: "BOOL"},
			{Name: "PGainInv", DataType: "BOOL"},
			{Name: "IGainInv", DataType: "BOOL"},
			{Name: "DGainInv", DataType: "BOOL"},
			{Name: "ZCDeadbandInv", DataType: "BOOL"},
			{Name: "PVDeadbandInv", DataType: "BOOL"},
			{Name: "PVROCLimitsInv", DataType: "BOOL"},
			{Name: "DevHLLimitsInv", DataType: "BOOL"},
			{Name: "DevDeadbandInv", DataType: "BOOL"},
			{Name: "AtuneDataInv", DataType: "BOOL"},
			{Name: "TimingModeInv", DataType: "BOOL"},
			{Name: "RTSMissed", DataType: "BOOL"},
			{Name: "RTSTimeInv", DataType: "BOOL"},
			{Name: "RTSTimeStampInv", DataType: "BOOL"},
			{Name: "DeltaTInv", DataType: "BOOL"},
		},
	}.AsType(*tl)
	if err != nil {
		return fmt.Errorf("could not register ControlLogix PID_ENHANCED because %w", err)
	}
	*tl = append(*tl, pidEnhanced)

	// The MESSAGE type has some data, but this will add it as `struct{}` since we don't use it
	*tl = append(*tl, structType{safeGoName: safeGoName{"MESSAGE", "MESSAGE"}})

	return nil
}

func (mb Member) AsNamedType(knownTypes TypeList) (NamedType, error) {
	if mb.DataType == "BOOL" && mb.Dimension > 1 {
		// When the BOOL type is stored in arrays, it appears to use 32-bit storage.
		// Maybe byte-sized access would also work. 🤷
		if mb.Dimension%32 != 0 {
			// I've only seen examples with multiples of 32.
			return NamedType{}, fmt.Errorf("%w is a BOOL array, but not a multiple of 32 (%d)", ErrUnknownType, mb.Dimension)
		}
		return knownTypes.newNamedType(mb.Name, "UDINT", []int{mb.Dimension / 32})
	}

	return knownTypes.newNamedType(mb.Name, mb.DataType, []int{mb.Dimension})
}

func (par Parameter) AsNamedType(knownTypes TypeList) (NamedType, error) {
	return knownTypes.newNamedType(par.Name, par.DataType, nil)
}

func newNamedType(name string, ty Type) (NamedType, error) {
	nt := NamedType{
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
	return nt, nil
}

// SetAsProgram changes the PlcName to indicate this is a program tag.
func (nt *NamedType) SetAsProgram() {
	if nt.PlcName == "" {
		nt.PlcName = nt.GoName
	}
	nt.PlcName = "Program:" + nt.PlcName
}

func (tl TypeList) newNamedType(name, dataType string, dims []int) (NamedType, error) {
	var nt NamedType
	for _, ty := range tl {
		if ty.PlcName() == dataType {
			var err error
			nt, err = newNamedType(name, ty)
			if err != nil {
				return NamedType{}, err
			}
			break
		}
	}
	if nt.Type == nil {
		return NamedType{}, ErrUnknownType
	}

	if len(dims) == 0 || len(dims) == 1 && dims[0] <= 1 { // not an array
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

func (aoi AddOnInstrDef) AsType(knownTypes TypeList) (Type, error) {
	sti, err := newStructType(aoi.Name, nil)
	if err != nil {
		return nil, err
	}
	for _, param := range aoi.Parameters {
		if param.DataType == "BIT" {
			// Not yet implemented, so ignore it
			continue
		}
		nm, err := param.AsNamedType(knownTypes)
		if err != nil {
			if errors.Is(err, ErrUnknownType) {
				err = errUnknownTypeSpecific{aoi.Name, param.DataType}
			}
			return nil, err
		}
		sti.members = append(sti.members, nm)
	}
	if len(sti.members) == 0 {
		return nil, fmt.Errorf("Add-on Instruction '%s' produced no parameters (%d)", aoi.Name, len(aoi.Parameters))
	}
	return sti, nil
}

func parseStruct(name string, membs []Member, knownTypes TypeList) (Type, error) {
	sti, err := newStructType(name, nil)
	if err != nil {
		return nil, err
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

	return newStringType(name, memb[1].Dimension)
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
	safeGoName
	members []NamedType
}

func (sti structType) GoTypeString() string {
	strs := make([]string, len(sti.members))
	for i, member := range sti.members {
		tagSuffix := ""
		if member.PlcName != "" {
			tagSuffix = fmt.Sprintf(" `%s:\"%s\"`", plc.TagPrefix, member.PlcName)
		}
		strs[i] = fmt.Sprintf("\n\t%s%s", NamedTypeDeclaration(member), tagSuffix)
	}
	return fmt.Sprintf("struct {%s\n}", strings.Join(strs, ""))
}
func newStructType(name string, members []NamedType) (structType, error) {
	if members == nil {
		members = []NamedType{}
	}
	sgn, err := newSafeGoName(name)
	if err != nil {
		return structType{}, err
	}
	return structType{
		safeGoName: sgn,
		members:    members,
	}, nil
}

type stringType struct {
	safeGoName
	count int
}

func (sty stringType) GoTypeString() string {
	return fmt.Sprintf("struct {Len int16; Data [%d]int8}", sty.count)
}
func newStringType(name string, count int) (stringType, error) {
	sgn, err := newSafeGoName(name)
	if err != nil {
		return stringType{}, err
	}
	return stringType{
		safeGoName: sgn,
		count:      count,
	}, nil
}

type safeGoName struct {
	goN, plcN string
}

func (sgn safeGoName) PlcName() string { return sgn.plcN }
func (sgn safeGoName) GoName() string  { return sgn.goN }
func newSafeGoName(name string) (safeGoName, error) {
	sgn := safeGoName{goN: name, plcN: name}
	if !isPublicGoIdentifier(name) {
		sgn.goN = makeValidIdentifier(sgn.goN)
		if sgn.goN == "" {
			return safeGoName{}, fmt.Errorf("couldn't create safe go name for '%s'", name)
		}
	}
	return sgn, nil
}

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

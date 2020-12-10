package plc

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Tag struct {
	name string
	TagType
	elementSize uint16
	dimensions  []int
}

func (tag *Tag) Name() string {
	return tag.name
}

func (tag *Tag) addDimension(dim int) {
	if dim <= 0 {
		return
	}
	tag.dimensions = append(tag.dimensions, dim)
}

func (tag Tag) String() string {
	name := fmt.Sprintf("%s{%v}", tag.name, tag.TagType)

	if len(tag.dimensions) == 0 {
		return name
	}

	strs := make([]string, len(tag.dimensions))
	for i, v := range tag.dimensions {
		strs[i] = strconv.Itoa(v)
	}
	return name + "[" + strings.Join(strs, ",") + "]"
}

func (tag Tag) ElemCount() int {
	count := 1
	for _, dim := range tag.dimensions {
		if dim != 0 {
			count *= dim
		}
	}
	return count
}

// ParseQualifiedTagName consumes a tag name containing
// zero or more qualifications (ie. a field name or an
// array index) and splits them into their respresentative
// parts.
//
// From libplctag (we are ignoring bit_seg)
/*
 * The EBNF is:
 *
 * tag ::= SYMBOLIC_SEG ( tag_seg )* ( bit_seg )?
 *
 * tag_seg ::= '.' SYMBOLIC_SEG
 *             '[' array_seg ']'
 *
 * bit_seg ::= '.' [0-9]+
 *
 * array_seg ::= NUMERIC_SEQ ( ',' NUMERIC_SEQ )*
 *
 * SYMBOLIC_SEG ::= [a-zA-Z]([a-zA-Z0-9_]*)
 *
 * NUMERIC_SEG ::= [0-9]+
 *
 */
func ParseQualifiedTagName(qtn string) ([]string, error) {
	var ret []string
	i := 0

	alpha := "abcdefhijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	num := "0123456789"
	alphanum := alpha + num + "_"

	/* SYMBOLIC_SEG := [a-zA-Z][a-zA-Z0-9_]* */
	parseSymbolicSegment := func() error {
		begin := i

		/* [a-zA-Z] */
		if i >= len(qtn) {
			return fmt.Errorf("Expected alphabetic character")
		}
		if unicode.IsSpace(rune(qtn[i])) {
			return fmt.Errorf("Expected alphabetic character, got whitespace")
		}
		if !bytes.ContainsAny([]byte{qtn[i]}, alpha) {
			return fmt.Errorf("symbolic sequence begins with a non-alphabetic character '%c'", qtn[i])
		}

		/* [a-zA-Z0-9_]* */
		for ; i < len(qtn); i++ {
			if unicode.IsSpace(rune(qtn[i])) {
				return fmt.Errorf("Expected alphanumeric character, got whitespace")
			}
			if !bytes.ContainsAny([]byte{qtn[i]}, alphanum) {
				break
			}
		}
		ret = append(ret, qtn[begin:i])
		return nil
	}

	/* NUMERIC_SEG := [:space:]* [0-9]+ [:space:]* */
	parseNumericSegment := func() error {
		var begin, end int

		/* [:space:]* */
		if i >= len(qtn) {
			return fmt.Errorf("expected number")
		}
		for ; i < len(qtn); i++ {
			if !unicode.IsSpace(rune(qtn[i])) {
				break
			}
		}

		/* [0-9]+ */
		begin = i

		/* [0-9]  */
		if i >= len(qtn) {
			return fmt.Errorf("expected number")
		}
		if !bytes.ContainsAny([]byte{qtn[i]}, num) {
			return fmt.Errorf("Expected digit, got '%c'", qtn[i])
		}
		i++
		/* [0-9]* */
		for ; i < len(qtn); i++ {
			if unicode.IsSpace(rune(qtn[i])) {
				break
			}
			if !bytes.ContainsAny([]byte{qtn[i]}, num) {
				break
			}
		}
		end = i

		/* [:space:]* */
		for ; i < len(qtn); i++ {
			if !unicode.IsSpace(rune(qtn[i])) {
				break
			}
		}

		asUint64, err := strconv.ParseUint(qtn[begin:end], 10, 32)
		if err != nil {
			return fmt.Errorf("Invalid array index '%s'", qtn[begin:end])
		}
		ret = append(ret, fmt.Sprintf("%d", asUint64))
		return nil
	}

	/* array_seg ::= numeric_seg ( ',' numeric_seg )* */
	parseArraySegment := func() error {
		if err := parseNumericSegment(); err != nil {
			return err
		}
		for i < len(qtn) {
			if qtn[i] != ',' {
				return nil
			}
			i++
			if err := parseNumericSegment(); err != nil {
				return err
			}
		}
		return nil
	}

	/* tag_seg ::= '.' SYMBOLIC_SEG | '[' array_seg ']' */
	parseTagSegment := func() error {
		if i >= len(qtn) {
			return fmt.Errorf("expected '.' or '['")
		}
		switch qtn[i] {
		case '.':
			i++
			return parseSymbolicSegment()
		case '[':
			i++
			if err := parseArraySegment(); err != nil {
				return err
			}
			if i >= len(qtn) {
				return fmt.Errorf("expected ']'")
			}
			if qtn[i] != ']' {
				return fmt.Errorf("expected ']'; got '%c'", qtn[i])
			}
			i++
		default:
			return fmt.Errorf("expected '.' or '['; got '%c'", qtn[i])
		}
		return nil
	}

	/* Check position-independent invariants: the tagname must be nonempty and
	 * must only contain alphanumeric characters.
	 */
	if qtn == "" {
		return nil, fmt.Errorf("Empty tagname")
	}
	for i, c := range qtn {
		if c > unicode.MaxASCII {
			return nil, fmt.Errorf("Non-ASCII character (codepoint %d) at index %d", int(c), i)
		}
	}

	/* tag ::= SYMBOLIC_SEG ( tag_seg )* */
	if err := parseSymbolicSegment(); err != nil {
		return nil, err
	}
	for i < len(qtn) {
		if err := parseTagSegment(); err != nil {
			return nil, err
		}
	}

	return ret, nil
}

type TagType uint16

func (tt TagType) String() string {
	if name, ok := tagTypeNames[tt]; ok {
		return name
	}
	if rtype, ok := tagTypes[tt]; ok {
		return fmt.Sprintf("%v", rtype)
	}
	return fmt.Sprintf("%04X", uint16(tt))
}

// NewInstance returns a pointer to a new variable of the TagType.
// This can be passed directly into ReadTag to read the value.
func (tt TagType) NewInstance() (interface{}, error) {
	rtyp, ok := tagTypes[tt]
	if !ok {
		return nil, fmt.Errorf("no instance of %s has been registered", tt.String())
	}

	return reflect.New(rtyp).Interface(), nil
}

func (tt TagType) HasName() bool {
	_, ok := tagTypeNames[tt]
	return ok
}

var tagTypeNames = map[TagType]string{
	0x00C1: "BOOL",
	0x00C2: "SINT",
	0x00C3: "INT",
	0x00C4: "DINT",
	0x00C5: "LINT",
	0x00C6: "USINT",
	0x00C7: "UINT",
	0x00C8: "UDINT",
	0x00C9: "ULINT",
	0x00CA: "REAL",
	0x00CB: "LREAL",
	0x00CC: "STIME",
	0x00CD: "DATE",
	0x00CE: "TIME_OF_DAY",
	0x00CF: "DATE_AND_TIME",
	0x00D0: "STRING",
	0x00D1: "BYTE",
	0x00D2: "WORD",
	0x00D3: "DWORD",
	0x00D4: "LWORD",
	0x00D5: "STRING2",
	0x00D6: "FTIME",
	0x00D7: "LTIME",
	0x00D8: "ITIME",
	0x00D9: "STRINGN",
	0x00DA: "SHORT_STRING",
	0x00DB: "TIME",
	0x00DC: "EPATH",
	0x00DD: "ENGUNIT",
	0x20C3: "INT",
	0x20C4: "DINT",
}

// RegisterTagTypeName registers the provided string as the name of the provided TagType.
// It returns an error if the TagType has already been registered with a different string,
// but it does not check if the string is unique.
func RegisterTagTypeName(tt TagType, name string) error {
	if name == "" {
		return fmt.Errorf("Cannot register empty string for TagType{%04X}", tt)
	}

	prevName, ok := tagTypeNames[tt]
	if ok { // tt is already registered
		if prevName == name {
			return nil // same string; do nothing
		}
		return fmt.Errorf("Cannot register string '%s' for TagType{%04X} with '%s' already registered", name, tt, prevName)
	}
	tagTypeNames[tt] = name
	return nil
}

func (tt TagType) CanBeInstantiated() bool {
	_, ok := tagTypes[tt]
	return ok
}

var tagTypes = map[TagType]reflect.Type{
	0x00C1: reflect.TypeOf(false),
	0x00C2: reflect.TypeOf(int8(0)),
	0x00C3: reflect.TypeOf(int16(0)),
	0x00C4: reflect.TypeOf(int32(0)),
	0x00C5: reflect.TypeOf(int64(0)),
	0x00C6: reflect.TypeOf(uint8(0)),
	0x00C7: reflect.TypeOf(uint16(0)),
	0x00C8: reflect.TypeOf(uint32(0)),
	0x00C9: reflect.TypeOf(uint64(0)),
	0x00CA: reflect.TypeOf(float32(0)),
	0x00CB: reflect.TypeOf(float64(0)),
	// 0x00CC: reflect.TypeOf(SynchronousTime{}), // unsupported
	// 0x00CD: reflect.TypeOf(Date{}), // unsupported
	// 0x00CE: reflect.TypeOf(TimeOfDay{}), // unsupported
	// 0x00CF: reflect.TypeOf(DateTime{}), // unsupported
	0x00D0: reflect.TypeOf(string("")),
	0x00D1: reflect.TypeOf(byte(0)),   // might not be correct; might be an array
	0x00D2: reflect.TypeOf(uint16(0)), // might not be correct; might be an array
	0x00D3: reflect.TypeOf(uint32(0)), // might not be correct; might be an array
	0x00D4: reflect.TypeOf(uint64(0)), // might not be correct; might be an array
	// 0x00D5: reflect.TypeOf(String2Byte{}), //unsupported
	// 0x00D6: reflect.TypeOf(DurationHigh{}), //unsupported, but uses int32
	// 0x00D7: reflect.TypeOf(DurationLong{}), //unsupported, but uses int64
	// 0x00D8: reflect.TypeOf(DurationShort{}), //unsupported
	// 0x00D9: reflect.TypeOf(CharStringNBytesPerCharacter{}), //unsupported
	// 0x00DA: reflect.TypeOf(CharacterSTringWithLengthByte{}), //unsupported
	// 0x00DB: reflect.TypeOf(Duration_ms{}), //unsupported
	// 0x00DC: reflect.TypeOf(CipPathSegments{}), //unsupported
	// 0x00DD: reflect.TypeOf(EngineeringUnits{}), //unsupported
	0x20C3: reflect.TypeOf(int32(0)), // might actually be unsigned or 16 bits?
	0x20C4: reflect.TypeOf(int32(0)),
}

// RegisterTagType registers the provided variable as the type of the provided TagType.
// It returns a warning if the TagType has already been registered with a different variable,
// but it does not check if the string is unique.
// This also registers the TagType name if there isn't already one.
func RegisterTagType(tt TagType, instanceOfType interface{}) error {
	if instanceOfType == nil {
		return fmt.Errorf("Cannot register nil for TagType{%04X}", tt)
	}

	newType := reflect.TypeOf(instanceOfType)

	prevType, ok := tagTypes[tt]
	if ok { // tt is already registered
		if prevType == newType {
			return nil // same type; do nothing
		}
		return fmt.Errorf("Cannot register type '%v' for TagType{%04X} with '%v' already registered", newType, tt, prevType)
	}
	tagTypes[tt] = newType

	return nil
}

package plc

/*
#include <stdint.h>
*/
import "C"
import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const SystemTagBit = 0x1000
const TagDimensionMask = 0x6000

type Tag struct {
	name        string
	tagType     C.uint16_t
	elementSize C.uint16_t
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
	name := fmt.Sprintf("%s{%04X}", tag.name, int(tag.tagType))

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
 * array_seg ::= NUMERIC_SEG ( ',' NUMERIC_SEG )*
 *
 * SYMBOLIC_SEG ::= [a-zA-Z]([a-zA-Z0-9_]*)
 *
 * NUMERIC_SEG ::= [0-9]+
 *
 */
func ParseQualifiedTagName(qtn string) ([]string, error) {
	var ret []string

	if qtn == "" {
		return nil, fmt.Errorf("Empty tagname supplied")
	}
	for i, c := range qtn {
		if unicode.IsSpace(c) {
			return nil, fmt.Errorf("Whitespace character at index %d", i)
		}
		if c < 32 || c > unicode.MaxASCII {
			return nil, fmt.Errorf("Non-ASCII character (codepoint %d) at index %d", int(c), i)
		}
	}

	fields := strings.Split(qtn, ".")
	for i, f := range fields {
		if f == "" {
			return nil, fmt.Errorf("Field #%d: Empty tagname supplied", i+1)
		}
		if bytes.ContainsAny([]byte{f[0]}, "0123456789_") {
			return nil, fmt.Errorf("Field #%d: Field begins with a non-alpha character '%c'", i+1, f[0])
		}
		openBracketIdx := strings.Index(f, "[")
		if openBracketIdx == -1 {
			// No '['; ensure the rest of the field has no unmatched ']'
			if strings.Index(f, "]") == -1 {
				ret = append(ret, f)
			} else {
				return nil, fmt.Errorf("Field #%d: ']' without '['", i+1)
			}
		} else if openBracketIdx == 0 {
			// The field begins with an open bracket; we're missing the field name
			return nil, fmt.Errorf("Field #%d: '[' without array identifier", i+1)
		} else {
			arrayName := f[:openBracketIdx]
			ret = append(ret, arrayName)
			arrayIndices := strings.Split(f[openBracketIdx+1:], "[")
			for _, a := range arrayIndices {
				// We've gotten as far as "FieldName["; do we have a matching ']'
				// and a valid unsigned int for the index?
				closing := strings.Index(a, "]")
				if closing == -1 {
					return nil, fmt.Errorf("Field #%d: '[' without ']'", i+1)
				}
				arrayContents := a[:closing]
				for _, arrayIdx := range strings.Split(arrayContents, ",") {
					if _, err := strconv.ParseUint(arrayIdx, 10, 32); err != nil {
						return nil, fmt.Errorf("Field #%d: Invalid array index %v: %v", i+1, arrayIdx, err)
					}
					ret = append(ret, arrayIdx)
				}
			}
		}
	}
	return ret, nil
}

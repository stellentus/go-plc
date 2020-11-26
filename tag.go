package plc

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Tag struct {
	name        string
	tagType     uint16
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

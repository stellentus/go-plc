package plc

import (
	"strings"
	"testing"
)

var parserTests = []struct {
	in     string
	out    []string
	errMsg string
}{
	{"", nil, "Empty tagname"},
	{" ", nil, "Whitespace"},
	{"\t", nil, "Whitespace"},
	{"\n", nil, "Whitespace"},
	{"\b", nil, "Non-ASCII"},
	{"标志名称", nil, "Non-ASCII"},
	{"DUMMY_AQUA_TEST_0", []string{"DUMMY_AQUA_TEST_0"}, ""},
	{"DUMMY_AQUA_TEST_0.DUMMY_AQUA_TEST_1", []string{"DUMMY_AQUA_TEST_0", "DUMMY_AQUA_TEST_1"}, ""},
	{"DUMMY_AQUA_TEST_0..DUMMY_AQUA_TEST_1", nil, "Empty"},
	{"[0]", nil, "'[' without array identifier"},
	{"ARRAY[0]", []string{"ARRAY", "0"}, ""},
	{"ARRAY[foo]", nil, "Invalid array index"},
	{"ARRAY[-1]", nil, "Invalid array index"},
	{"ARRAY[0", nil, "'[' without ']'"},
	{"ARRAY[[0", nil, "'[' without ']'"},
	{"ARRAY[0][", nil, "'[' without ']'"},
	{"ARRAY0]", nil, "']' without '['"},
	{"DUMMY_AQUA_TEST.[0]", nil, "'[' without array identifier"},
	{"ARRAY[0][1][2]", []string{"ARRAY", "0", "1", "2"}, ""},
	{"Field.Array[42].Member[16]", []string{"Field", "Array", "42", "Member", "16"}, ""},
}

func compareStrSlices(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func TestParser(t *testing.T) {
	for _, test := range parserTests {
		out, err := ParseQualifiedTagName(test.in)
		if !compareStrSlices(out, test.out) {
			t.Errorf("ParseQualifiedTagName(\"%v\"): Return value: Got %v, expected %v", test.in, out, test.out)
		}
		if err == nil && test.errMsg != "" {
			t.Errorf("ParseQualifiedTagName(\"%v\"): Error value: Got nil error, expected error containing %v", test.in, test.errMsg)
		}
		if err != nil && test.errMsg == "" {
			t.Errorf("ParseQualifiedTagName(\"%v\"): Return value: Got non-nil error %v, expected nil error", test.in, err)
		}
		if err != nil {
			if !strings.Contains(err.Error(), test.errMsg) {
				t.Errorf("ParseQualifiedTagName(\"%v\"): Error value: Got \"%v\", should contain \"%v\"", test.in, err.Error(), test.errMsg)
			}
		}
	}
}

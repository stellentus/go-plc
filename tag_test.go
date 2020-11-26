package plc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var parserTests = []struct {
	in     string
	out    []string
	errMsg string
}{
	{"", nil, "Empty tagname"},
	{" ", nil, "whitespace"},
	{"\t", nil, "whitespace"},
	{"\n", nil, "whitespace"},
	{"\b", nil, "non-alphabetic character"},
	{"标志名称", nil, "Non-ASCII"},
	{"4UMMY_AQUA_TEST_0", nil, "begins with a non-alphabetic character '4'"},
	{"_UMMY_AQUA_TEST_0", nil, "begins with a non-alphabetic character '_'"},
	{"DUMMY_AQUA_TEST_0", []string{"DUMMY_AQUA_TEST_0"}, ""},
	{"DUMMY_AQUA_TEST_0.DUMMY_AQUA_TEST_1", []string{"DUMMY_AQUA_TEST_0", "DUMMY_AQUA_TEST_1"}, ""},
	{"DUMMY_AQUA_TEST_0..DUMMY_AQUA_TEST_1", nil, "begins with a non-alphabetic character '.'"},
	{"[0]", nil, "begins with a non-alphabetic character '['"},
	{"ARRAY[0]", []string{"ARRAY", "0"}, ""},
	{"ARRAY[foo]", nil, "Expected digit"},
	{"ARRAY[-1]", nil, "Expected digit"},
	{"ARRAY[0", nil, "expected ']'"},
	{"ARRAY[[0", nil, "Expected digit"},
	{"ARRAY[0][", nil, "expected number"},
	{"ARRAY0]", nil, "expected '.' or '['"},
	{"DUMMY_AQUA_TEST.[0]", nil, "non-alphabetic character '['"},
	{"ARRAY[]", nil, "Expected digit"},
	{"ARRAY[0,,2]", nil, "Expected digit"},
	{"ARRAY[0,foo,2]", nil, "Expected digit"},
	{"ARRAY[0][1][2]", []string{"ARRAY", "0", "1", "2"}, ""},
	{"ARRAY[0,1,2]", []string{"ARRAY", "0", "1", "2"}, ""},
	{"ARRAY[ 0 ,  1  , 2 ]", []string{"ARRAY", "0", "1", "2"}, ""},
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
			t.Errorf(`ParseQualifiedTagName("%v"): Return value: Got "%v", expected "%v"`, test.in, out, test.out)
		}
		if err == nil && test.errMsg != "" {
			t.Errorf(`"ParseQualifiedTagName("%v"): Error value: Got nil error, expected error containing "%v"`, test.in, test.errMsg)
		}
		if err != nil && test.errMsg == "" {
			t.Errorf(`ParseQualifiedTagName("%v"): Return value: Got non-nil error %v, expected nil error`, test.in, err)
		}
		if err != nil {
			if !strings.Contains(err.Error(), test.errMsg) {
				t.Errorf(`ParseQualifiedTagName("%v"): Error value: Got "%v", should contain "%v"`, test.in, err.Error(), test.errMsg)
			}
		}
	}
}

func resetRegistration() {
	tagTypeNames = map[TagType]string{}
}

var testTagType = TagType(0x7738)
var testTagTypeName = "testTagType"

func TestTagTypeString(t *testing.T) {
	assert.Equal(t, "7738", testTagType.String())
}

func TestTagTypeHasNoName(t *testing.T) {
	assert.False(t, testTagType.HasName(), "unregistered type has no name")
}

func TestRegisterTagType(t *testing.T) {
	resetRegistration()
	err := RegisterTagTypeName(testTagType, testTagTypeName)
	require.NoError(t, err)
	assert.Equal(t, testTagTypeName, testTagType.String(), "Name should be updated if registered")
}

func TestRegisterDuplicateTagType(t *testing.T) {
	resetRegistration()
	err := RegisterTagTypeName(testTagType, testTagTypeName)
	require.NoError(t, err)
	err = RegisterTagTypeName(testTagType, testTagTypeName)
	assert.NoError(t, err)
}

func TestRegisterTwoStringsOneTag(t *testing.T) {
	resetRegistration()
	err := RegisterTagTypeName(testTagType, testTagTypeName)
	require.NoError(t, err)
	err = RegisterTagTypeName(testTagType, testTagTypeName+"DifferentString")
	assert.Error(t, err)
}

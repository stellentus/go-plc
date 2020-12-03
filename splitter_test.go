package plc

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSplitReaderForTesting() (SplitReader, FakeReadWriter) {
	fakeRW := FakeReadWriter(map[string]interface{}{})
	return NewSplitReader(fakeRW), fakeRW
}

func newSplitWriterForTesting() (SplitWriter, FakeReadWriter) {
	fakeRW := FakeReadWriter(map[string]interface{}{})
	return NewSplitWriter(fakeRW), fakeRW
}

var manyTypesToTest = []interface{}{
	uint8(7),
	uint16(7),
	uint32(7),
	uint64(7),
	int8(7),
	int16(7),
	int32(7),
	int64(7),
	float32(7),
	float64(7),
}

func TestSplitReader(t *testing.T) {
	for _, tc := range manyTypesToTest {
		t.Run(reflect.TypeOf(tc).String(), func(tt *testing.T) {
			sr, fakeRW := newSplitReaderForTesting()
			fakeRW[testTagName] = tc

			// Create an actual variable of the type we want to test
			actual := reflect.New(reflect.TypeOf(tc)).Interface()
			require.Equal(tt, reflect.TypeOf(actual), reflect.PtrTo(reflect.TypeOf(tc)), "Created type must match desired type") // If this fails, it's a bug in test code, not the underlying code.

			// Now read the variable and make sure it is the same
			err := sr.ReadTag(testTagName, actual)
			require.NoError(tt, err)
			require.Equal(tt, tc, reflect.ValueOf(actual).Elem().Interface())
		})
	}
}

func TestSplitWriter(t *testing.T) {
	for _, tc := range manyTypesToTest {
		t.Run(reflect.TypeOf(tc).String(), func(tt *testing.T) {
			sr, fakeRW := newSplitWriterForTesting()

			// Now read the variable and make sure it is the same
			err := sr.WriteTag(testTagName, tc)
			require.NoError(tt, err)
			require.Equal(tt, tc, fakeRW[testTagName])
		})
	}
}

func TestSplitWriterWithPointer(t *testing.T) {
	val := int32(5)
	sr, fakeRW := newSplitWriterForTesting()

	// Now read the variable and make sure it is the same
	err := sr.WriteTag(testTagName, &val)
	require.NoError(t, err)
	require.Equal(t, val, fakeRW[testTagName])
}

// TestSplitReaderError is sort of testing the FakeReadWriter, not so much the SplitReader.
func TestSplitReaderError(t *testing.T) {
	for _, tc := range manyTypesToTest {
		t.Run(reflect.TypeOf(tc).String(), func(tt *testing.T) {
			sr, fakeRW := newSplitReaderForTesting()
			fakeRW[testTagName] = int(7)

			// Create an actual variable of the type we want to test
			actual := reflect.New(reflect.TypeOf(tc)).Interface()
			require.Equal(tt, reflect.TypeOf(actual), reflect.PtrTo(reflect.TypeOf(tc)), "Created type must match desired type") // If this fails, it's a bug in test code, not the underlying code.

			// Read fails because the data type is different. Note int!=int32 or any other size.
			err := sr.ReadTag(testTagName, actual)
			require.Error(tt, err)
		})
	}
}

type testStructType struct {
	I        uint32
	MY_FLOAT float64
}
type recursionType struct {
	VAL         int8
	STRUCT_HERE testStructType
}
type structWithUnexported struct {
	BIG        uint64
	unexported int32
}
type structWithTags struct {
	BIG        uint64 `plctag:"myName,omitempty"`
	unexported int32  `plctag:"StillUnexported"`
}
type structIgnoringTag struct {
	BIG    uint64 `plctag:"-"`
	MEDIUM int32  `plctag:""`
}
type structIgnoringTagDashComma struct {
	MEDIUM int32 `plctag:""`
	SMALL  int8  `plctag:"-,"`
}

func TestSplitReadStruct(t *testing.T) {
	expected := testStructType{7, 3.14}

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".I"] = expected.I
	fakeRW[testTagName+".MY_FLOAT"] = expected.MY_FLOAT

	actual := testStructType{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReadStructInStruct(t *testing.T) {
	expected := recursionType{-5, testStructType{7, 3.14}}

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".VAL"] = expected.VAL
	fakeRW[testTagName+".STRUCT_HERE.I"] = expected.STRUCT_HERE.I
	fakeRW[testTagName+".STRUCT_HERE.MY_FLOAT"] = expected.STRUCT_HERE.MY_FLOAT

	actual := recursionType{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReadStructUnexported(t *testing.T) {
	expected := structWithUnexported{BIG: 12} // Don't bother filling 'unexported' because it won't be set

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".BIG"] = expected.BIG
	// Since we don't save ".unexported", if there's an attempt to read it, an error will occur

	actual := structWithUnexported{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReadStructTag(t *testing.T) {
	expected := structWithTags{BIG: 7}

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".myName"] = expected.BIG

	actual := structWithTags{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReadStructTagIgnored(t *testing.T) {
	expected := structIgnoringTag{MEDIUM: 7}

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".MEDIUM"] = expected.MEDIUM

	actual := structIgnoringTag{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReadStructTagIgnoredDashComma(t *testing.T) {
	expected := structIgnoringTagDashComma{MEDIUM: 7}

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".MEDIUM"] = expected.MEDIUM

	actual := structIgnoringTagDashComma{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitWriteStruct(t *testing.T) {
	expected := testStructType{7, 3.14}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, expected.I, fakeRW[testTagName+".I"])
	assert.Equal(t, expected.MY_FLOAT, fakeRW[testTagName+".MY_FLOAT"])
}

func TestSplitWriteStructPointer(t *testing.T) {
	expected := testStructType{7, 3.14}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, &expected)
	require.NoError(t, err)

	assert.Equal(t, expected.I, fakeRW[testTagName+".I"])
	assert.Equal(t, expected.MY_FLOAT, fakeRW[testTagName+".MY_FLOAT"])
}

func TestSplitWriteStructInStruct(t *testing.T) {
	expected := recursionType{-5, testStructType{7, 3.14}}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, expected.VAL, fakeRW[testTagName+".VAL"])
	assert.Equal(t, expected.STRUCT_HERE.I, fakeRW[testTagName+".STRUCT_HERE.I"])
	assert.Equal(t, expected.STRUCT_HERE.MY_FLOAT, fakeRW[testTagName+".STRUCT_HERE.MY_FLOAT"])
}

func TestSplitWriteStructUnexported(t *testing.T) {
	expected := structWithUnexported{BIG: 12, unexported: 57}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, 1, len(fakeRW), "Only 1 value should be written")
	assert.Equal(t, expected.BIG, fakeRW[testTagName+".BIG"])
}

func TestSplitWriteStructTag(t *testing.T) {
	expected := structWithTags{BIG: 7, unexported: 57}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, 1, len(fakeRW), "Only 1 value should be written")
	assert.Equal(t, expected.BIG, fakeRW[testTagName+".myName"])
}

func TestSplitWriteOmitEmpty(t *testing.T) {
	expected := structWithTags{BIG: 0, unexported: 57}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, FakeReadWriter{}, fakeRW, "No values should be written")
}

func TestSplitWriteStructTagIgnored(t *testing.T) {
	expected := structIgnoringTag{MEDIUM: 7}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, 1, len(fakeRW), "Only 1 value should be written")
	assert.Equal(t, expected.MEDIUM, fakeRW[testTagName+".MEDIUM"])
}

func TestSplitWriteStructTagIgnoredDashComma(t *testing.T) {
	expected := structIgnoringTagDashComma{MEDIUM: 7}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, 1, len(fakeRW), "Only 1 value should be written")
	assert.Equal(t, expected.MEDIUM, fakeRW[testTagName+".MEDIUM"])
}

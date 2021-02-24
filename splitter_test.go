package plc

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitReaderParallel(t *testing.T) {
	const expected = uint8(7)
	fakeRW := FakeReadWriter(map[string]interface{}{})
	fakeRW[testTagName] = expected

	// Now read the variable and make sure it is the same
	var actual uint8
	err := NewSplitReaderParallel(fakeRW).ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReaderParallelIsParallel(t *testing.T) {
	numToTest := 8

	done := make(chan struct{})
	count := make(chan int, numToTest)

	sr := NewSplitReaderParallel(readerFunc(func(name string, value interface{}) error {
		var idx int
		_, err := fmt.Sscanf(name, testTagName+"[%d]", &idx)
		count <- idx // indicate this index executed

		<-done // Block until this channel is closed
		return err
	}))

	go func() {
		data := make([]int, numToTest) // We don't actually modify this data.
		err := sr.ReadTag(testTagName, &data)
		assert.NoError(t, err)
	}()

	receivedIndices := uint32(0) // this limits the max number of parallel tests to 32
	numReceived := 0

outerLoop:
	for {
		select {
		case <-time.After(50 * time.Millisecond):
			assert.Fail(t, "Timeout in reading from SplitReaderParallel")
			break outerLoop
		case idx := <-count:
			receivedIndices |= 1 << uint32(idx)
			numReceived++
			if numReceived >= numToTest {
				break outerLoop
			}
		}
	}
	close(done)

	assert.Equal(t, uint32(1<<numToTest)-1, receivedIndices, "Not all expected SplitReader occurred in parallel")
}

func TestSplitReaderParallelMulti(t *testing.T) {
	const expected = uint8(7)
	fakeRW := FakeReadWriter(map[string]interface{}{})
	sr := NewSplitReaderParallel(fakeRW)

	for _, tc := range manyTypesToTest {
		fakeRW[testTagName+reflect.TypeOf(tc).String()] = tc
	}

	for _, tc := range manyTypesToTest {
		// Create an actual variable of the type we want to test
		actual := reflect.New(reflect.TypeOf(tc)).Interface()
		require.Equal(t, reflect.TypeOf(actual), reflect.PtrTo(reflect.TypeOf(tc)), "Created type must match desired type") // If this fails, it's a bug in test code, not the underlying code.

		// Now read the variable and make sure it is the same
		err := sr.ReadTag(testTagName+reflect.TypeOf(tc).String(), actual)
		require.NoError(t, err)
		require.Equal(t, tc, reflect.ValueOf(actual).Elem().Interface())
	}
}

func TestSplitReaderParallelMany(t *testing.T) {
	fakeRW := FakeReadWriter(map[string]interface{}{})
	sr := NewSplitReaderParallel(fakeRW)

	const testLength = 512
	expected := make([]int, testLength)
	for i := 0; i < testLength; i++ {
		fakeRW[testTagName+"["+strconv.Itoa(i)+"]"] = i
		expected[i] = i
	}

	actual := make([]int, testLength)
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func TestSplitReaderParallelError(t *testing.T) {
	fakeRW := FakeReadWriter(map[string]interface{}{})
	fakeRW[testTagName] = int(7)

	// Read fails because the data type is different. Note int!=int32 or any other size.
	var actual uint8
	err := NewSplitReaderParallel(fakeRW).ReadTag(testTagName, &actual)
	require.Error(t, err)
}

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
type structWithPointer struct {
	POINT *uint64 `plctag:",omitempty"`
}
type structWithArray struct {
	Vals [2]int
}
type structWithSlice struct {
	Vals []int
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

func TestSplitReadStructTagWithPointer(t *testing.T) {
	expected := uint64(14)

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+".POINT"] = expected

	actual := structWithPointer{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, *actual.POINT)
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

func TestSplitWriteStructTagWithPointer(t *testing.T) {
	val := uint64(0) // Should be set, even though it's 0 and omitempty
	expected := structWithPointer{POINT: &val}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, 1, len(fakeRW), "Only 1 value should be written")
	assert.Equal(t, *expected.POINT, fakeRW[testTagName+".POINT"])
}

func TestSplitWriteStructTagWithNilPointer(t *testing.T) {
	expected := structWithPointer{} // No value set
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, 0, len(fakeRW), "Omitempty pointer should not be written")
}

var expectedArray = [2]int{5, 83}

func TestSplitReadSlice(t *testing.T) {
	tests := []struct {
		expected interface{}
		actual   []int
		message  string
	}{
		{expectedArray[:], make([]int, 2), "FullSlice"},
		{expectedArray[0:1], make([]int, 1), "PartialSlice"},
	}

	for _, test := range tests {
		t.Run(test.message, func(tt *testing.T) {
			sr, fakeRW := newSplitReaderForTesting()
			fakeRW[testTagName+"[0]"] = expectedArray[0]
			fakeRW[testTagName+"[1]"] = expectedArray[1]

			err := sr.ReadTag(testTagName, &test.actual)
			require.NoError(t, err)
			require.Equal(t, test.expected, test.actual)
		})
	}
}

func TestSplitReadArray(t *testing.T) {
	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+"[0]"] = expectedArray[0]
	fakeRW[testTagName+"[1]"] = expectedArray[1]

	actual := [2]int{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expectedArray, actual)
}

func TestSplitReadArrayOfStruct(t *testing.T) {
	expected := [2]testStructType{
		testStructType{7, 3.14},
		testStructType{83, .11},
	}

	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName+"[0].I"] = expected[0].I
	fakeRW[testTagName+"[0].MY_FLOAT"] = expected[0].MY_FLOAT
	fakeRW[testTagName+"[1].I"] = expected[1].I
	fakeRW[testTagName+"[1].MY_FLOAT"] = expected[1].MY_FLOAT

	actual := [2]testStructType{}
	err := sr.ReadTag(testTagName, &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

// indirect assumes val is a pointer to something, and it returns "something".
func indirect(val interface{}) interface{} {
	return reflect.ValueOf(val).Elem().Interface()
}

func TestSplitReadArrayInStruct(t *testing.T) {
	tests := []struct {
		expected interface{}
		actual   interface{}
		message  string
	}{
		{structWithArray{Vals: expectedArray}, &structWithArray{}, "ArrayInStruct"},
		{structWithSlice{Vals: expectedArray[:]}, &structWithSlice{Vals: make([]int, 2)}, "SliceInStruct"},
		{structWithSlice{Vals: nil}, &structWithSlice{Vals: nil}, "NilSliceInStruct"},
		{structWithSlice{Vals: []int{}}, &structWithSlice{Vals: []int{}}, "EmptySliceInStruct"},
	}

	for _, test := range tests {
		t.Run(test.message, func(tt *testing.T) {
			sr, fakeRW := newSplitReaderForTesting()
			fakeRW[testTagName+".Vals[0]"] = expectedArray[0]
			fakeRW[testTagName+".Vals[1]"] = expectedArray[1]

			err := sr.ReadTag(testTagName, test.actual)
			require.NoError(t, err)
			require.Equal(t, test.expected, indirect(test.actual))
		})
	}
}

func TestSplitWriteArray(t *testing.T) {
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expectedArray)
	require.NoError(t, err)

	assert.Equal(t, expectedArray[0], fakeRW[testTagName+"[0]"])
	assert.Equal(t, expectedArray[1], fakeRW[testTagName+"[1]"])
}

func TestSplitWriteSlice(t *testing.T) {
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expectedArray[:])
	require.NoError(t, err)

	assert.Equal(t, expectedArray[0], fakeRW[testTagName+"[0]"])
	assert.Equal(t, expectedArray[1], fakeRW[testTagName+"[1]"])
}

func TestSplitWriteArrayOfStruct(t *testing.T) {
	expected := [2]testStructType{
		testStructType{7, 3.14},
		testStructType{83, .11},
	}
	sw, fakeRW := newSplitWriterForTesting()

	err := sw.WriteTag(testTagName, expected)
	require.NoError(t, err)

	assert.Equal(t, expected[0].I, fakeRW[testTagName+"[0].I"])
	assert.Equal(t, expected[0].MY_FLOAT, fakeRW[testTagName+"[0].MY_FLOAT"])
	assert.Equal(t, expected[1].I, fakeRW[testTagName+"[1].I"])
	assert.Equal(t, expected[1].MY_FLOAT, fakeRW[testTagName+"[1].MY_FLOAT"])
}

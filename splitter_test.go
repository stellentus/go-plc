package plc

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func newSplitReaderForTesting() (SplitReader, FakeReadWriter) {
	fakeRW := FakeReadWriter(map[string]interface{}{})
	return NewSplitReader(fakeRW), fakeRW
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

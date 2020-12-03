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

func TestSplitReader(t *testing.T) {
	testCases := []interface{}{
		uint8(7),
	}

	for _, tc := range testCases {
		t.Run(reflect.TypeOf(tc).String(), func(tt *testing.T) {
			sr, fakeRW := newSplitReaderForTesting()
			fakeRW[testTagName] = tc

			// Create an actual variable of the type we want to test
			actual := reflect.New(reflect.TypeOf(tc)).Interface()
			require.Equal(tt, reflect.TypeOf(actual), reflect.PtrTo(reflect.TypeOf(tc)), "Created type must match desired type")

			// Now read the variable and make sure it is the same
			err := sr.ReadTag(testTagName, actual)
			require.NoError(tt, err)
			require.Equal(tt, tc, reflect.ValueOf(actual).Elem().Interface())
		})
	}
}

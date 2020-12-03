package plc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func newSplitReaderForTesting() (SplitReader, FakeReadWriter) {
	fakeRW := FakeReadWriter(map[string]interface{}{})
	return NewSplitReader(fakeRW), fakeRW
}

func TestSplitReader(t *testing.T) {
	sr, fakeRW := newSplitReaderForTesting()
	fakeRW[testTagName] = 7

	var actual int
	err := sr.ReadTag(testTagName, &actual)
	assert.NoError(t, err)
	assert.Equal(t, 7, actual)
}

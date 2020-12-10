package l5x

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testFilePath = "test.L5X"

func TestParse(t *testing.T) {
	_, err := ParseFromFile(testFilePath)
	require.NoError(t, err)
}

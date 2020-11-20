package plc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCacheForTesting() (*Cache, DeviceFake) {
	devFake := DeviceFake(map[string]interface{}{})
	return NewCache(devFake), devFake
}

func TestNewCache(t *testing.T) {
	cache, _ := newCacheForTesting()
	assert.NotNil(t, cache, "NewCache should return a non-nil object")
}

func TestCachePassThroughInt(t *testing.T) {
	cache, deviceFake := newCacheForTesting()
	deviceFake[testTagName] = 7

	var actual int
	err := cache.ReadTag(testTagName, &actual)
	assert.NoError(t, err)
	assert.Equal(t, 7, actual)
}

func TestCacheCachesInt(t *testing.T) {
	cache, deviceFake := newCacheForTesting()
	deviceFake[testTagName] = 7

	var unused int
	err := cache.ReadTag(testTagName, &unused)
	require.NoError(t, err)
	unused++ // increment the value to ensure we aren't just returning a pointer

	var actual int
	err = cache.ReadCachedTag(testTagName, &actual)
	assert.NoError(t, err)
	assert.Equal(t, 7, actual)
}

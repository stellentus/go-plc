package plc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncAddError(t *testing.T) {
	as := newAsync(nil)
	expected := errors.New("Test")
	as.AddError(expected)
	actual := as.Wait()
	assert.Equal(t, expected, actual, "Error should be propagated")
}

func TestAsyncAddTwoErrors(t *testing.T) {
	as := newAsync(nil)
	expected1 := errors.New("Test1")
	expected2 := errors.New("Test2")
	as.AddError(expected1)
	as.AddError(expected2)
	actual := as.Wait()
	if expected1 != actual && expected2 != actual {
		assert.Fail(t, "Error message was not either expected value")
	}
}

func TestNewAsync(t *testing.T) {
	name := "TESTNAME"
	value := int(8)
	expErr := errors.New("Test")
	as := newAsync(func(nm string, val interface{}) error {
		require.Equal(t, name, nm, "Async should pass the name unchanged")
		require.Equal(t, value, val, "Async should pass the value unchanged")
		return expErr
	})
	as.Add(name, value)
	err := as.Wait()
	require.Equal(t, expErr, err, "Async should pass the error unchanged")
}

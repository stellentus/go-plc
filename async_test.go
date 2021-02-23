package plc

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
		assert.Equal(t, name, nm, "Async should pass the name unchanged")
		assert.Equal(t, value, val, "Async should pass the value unchanged")
		return expErr
	})
	as.Add(name, value)
	err := as.Wait()
	assert.Equal(t, expErr, err, "Async should pass the error unchanged")
}

func TestAsyncMany(t *testing.T) {
	sendVals := make([]int, 8)
	receiveVals := make([]int, 8)
	for i := range sendVals {
		sendVals[i] = i
		receiveVals[i] = -1
	}

	as := newAsync(func(nm string, val interface{}) error {
		idx, err := strconv.Atoi(nm)
		receiveVals[idx] = val.(int)
		return err
	})

	for i := range sendVals {
		as.Add(strconv.Itoa(i), sendVals[i])
	}

	err := as.Wait()
	assert.NoError(t, err)
	assert.Equal(t, sendVals, receiveVals)
}

func TestAsyncManyParallel(t *testing.T) {
	numToTest := 8
	sendVals := make([]int, numToTest)
	receiveVals := make([]int, numToTest)
	for i := range sendVals {
		sendVals[i] = i
		receiveVals[i] = -1
	}

	done := make(chan struct{})
	count := make(chan int, numToTest)

	as := newAsync(func(nm string, val interface{}) error {
		idx, err := strconv.Atoi(nm)
		count <- idx // indicate this index executed
		receiveVals[idx] = val.(int)

		<-done // Block until this channel is closed
		return err
	})

	for i := range sendVals {
		as.Add(strconv.Itoa(i), sendVals[i])
	}

	receivedIndices := uint32(0) // this limits the max number of parallel tests to 32
	numReceived := 0

outerLoop:
	for {
		select {
		case <-time.After(50 * time.Millisecond):
			assert.Fail(t, "Timeout in reading from parallel async")
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

	err := as.Wait()
	assert.Equal(t, uint32(1<<numToTest)-1, receivedIndices, "Not all expected async occurred in parallel")
	assert.NoError(t, err)
	assert.Equal(t, sendVals, receiveVals)
}

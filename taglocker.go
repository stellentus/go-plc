package plc

import (
	"fmt"
	"sync"

	"github.com/dijkstracula/go-ilock"
)

// TagLockerOperationError is returned when one or more atomic operations on
// the tag locker in a "serialized" operation have failed. Since a "transaction"
// on the tag tree can yield multiple errors, a TagLockerFailure is the
// composition of one or more individual errors.
type TagLockerOperationError error

func lockError(e error) TagLockerOperationError {
	return fmt.Errorf("Error locking tag locker: %v", e)
}

func unlockError(e error) TagLockerOperationError {
	return fmt.Errorf("Error unlocking tag locker: %v", e)
}

func rLockError(e error) TagLockerOperationError {
	return fmt.Errorf("Error read-locking tag locker: %v", e)
}

func rUnlockError(e error) TagLockerOperationError {
	return fmt.Errorf("Error read-unlocking tag locker: %v", e)
}

// TagLocker is a plc.ReadWriter that wraps another ReadWriter, but gates
// concurrent accesses on grabbing read or write access on a tree of locks
// representing tag names (in the case of tree leaf nodes) and prefixes of tag
// names (in the case of tree frond nodes).
type TagLocker struct {
	downstream ReadWriter
	tagTree    *tagLockerNode
}

// ReadTag reads the given tag name from the downstream ReadWriter. If another
// thread is concurrently writing to this tag or a prefix of the tag, we will
// block until that thread has released its access.
func (tl *TagLocker) ReadTag(name string, value interface{}) (err error) {
	components, err := ParseQualifiedTagName(name)
	if err != nil {
		return
	}

	err = tl.tagTree.rLock(components)
	if err != nil {
		err = rLockError(err)
		return
	}

	defer func() {
		unlockErr := tl.tagTree.rUnlock(components)
		if unlockErr != nil {
			err = rUnlockError(unlockErr)
			return
		}
	}()

	err = tl.downstream.ReadTag(name, value)
	if err != nil {
		return
	}

	return
}

// WriteTag writes the given tag value to the downstream ReadWriter. Will block
// if another thread is reading this tag or a prefix of the tag.
func (tl *TagLocker) WriteTag(name string, value interface{}) (err error) {
	components, err := ParseQualifiedTagName(name)
	if err != nil {
		return
	}

	err = tl.tagTree.lock(components)
	if err != nil {
		err = lockError(err)
		return
	}

	defer func() {
		unlockErr := tl.tagTree.unlock(components)
		if unlockErr != nil {
			err = unlockError(unlockErr)
			return
		}
	}()

	err = tl.downstream.WriteTag(name, value)
	if err != nil {
		return
	}

	return
}

type tagLockerNode struct {
	mtx sync.RWMutex // Ensures mutual exclusion on the fields of the node.

	tagLock   *ilock.Mutex              // The logical lock that mutator threads will hold while reading and writing tags.
	component string                    // The component of the tag name.
	children  map[string]*tagLockerNode // All descendents of this node.
}

// NewTagLocker produces a new TagLocker.
func NewTagLocker(downstream ReadWriter) *TagLocker {
	return &TagLocker{
		downstream: downstream,
		tagTree:    newNode("/"),
	}
}

func newNode(component string) *tagLockerNode {
	return &tagLockerNode{
		tagLock:   ilock.New(),
		mtx:       sync.RWMutex{},
		component: component,
		children:  make(map[string]*tagLockerNode),
	}
}

// getOrCreate atomically returns the child of `tn` with the supplied
// component name; or, creates and inserts a child with that name.  In
// either case, the child in question is returned.
// Assumes that tn.mtx is _not_ held!
func (tn *tagLockerNode) getOrCreateChild(component string) *tagLockerNode {
	tn.mtx.RLock()
	child, ok := tn.children[component]

	// Do we already have a child component with the current component
	// name? If so, just recurse on that.
	if ok {
		tn.mtx.RUnlock()
		return child
	}

	// The child does not exist.  Upgrade to a writer lock in order
	// to insert a new child into our set of children.
	tn.mtx.RUnlock()
	tn.mtx.Lock()

	// Check again to see that nobody beat us to creating that child.
	child, ok = tn.children[component]
	if ok {
		// Lucky us!
		tn.mtx.Unlock()
		return child
	}

	// Okay, we have no choice but to create the child ourselves.
	child = newNode(component)
	tn.children[component] = child

	// Release our own lock and recurse on the child.
	tn.mtx.Unlock()
	return child
}

// lock traverses the slice of components, setting intention writer
// locks along the branch of the lock tree, until it reaches the final
// tag component.  There, it grabs an exclusive writer lock on that node.
func (tn *tagLockerNode) lock(components []string) error {

	// If we have no paths to traverse, lock ourselves!
	if len(components) == 0 {
		//fmt.Fprintf(os.Stderr, "XLock %v\n", tn.component)
		tn.tagLock.XLock()
		return nil
	}

	// Otherwise, we are only part of the way to the final component
	// to lock.  Take an intent read lock on this node.
	//fmt.Fprintf(os.Stderr, "IXLock %v\n", tn.component)
	tn.tagLock.IXLock()

	currentComp := components[0]
	remainingComp := components[1:]
	return tn.getOrCreateChild(currentComp).lock(remainingComp)
}

// rLock traverses the slice of components, setting intention reader
// locks along the branch of the lock tree, until it reaches the final
// tag component.  There, it grabs a reader lock on that node.
func (tn *tagLockerNode) rLock(components []string) error {
	// If we have no paths to traverse, lock ourselves!
	if len(components) == 0 {
		//fmt.Fprintf(os.Stderr, "SLock %v\n", tn.component)
		tn.tagLock.SLock()
		return nil
	}

	// Otherwise, we are only part of the way to the final component
	// to lock.  Take an intent read lock on this node.
	//fmt.Fprintf(os.Stderr, "ISLock %v\n", tn.component)
	tn.tagLock.ISLock()
	currentComp := components[0]
	remainingComp := components[1:]

	return tn.getOrCreateChild(currentComp).rLock(remainingComp)
}

// rUnlock unlocks a path that has already been previously
// locked for reader access.  It decrements the reader count on
// the final component in the path and removes a read intention
// on all other paths.
func (tn *tagLockerNode) rUnlock(components []string) error {
	// If we have no paths to traverse, unlock ourselves!
	if len(components) == 0 {
		//fmt.Fprintf(os.Stderr, "SUnLock %v\n", tn.component)
		tn.tagLock.SUnlock()
		return nil
	}

	currentComp := components[0]
	remainingComp := components[1:]

	defer func() {
		//fmt.Fprintf(os.Stderr, "ISUnLock %v\n", tn.component)
		tn.tagLock.ISUnlock()
	}()

	// Unlock our children first - we want to unlock in the opposite
	// order that we acquired the locks.
	tn.mtx.RLock()
	child, ok := tn.children[currentComp]
	tn.mtx.RUnlock()
	if !ok {
		return fmt.Errorf("missing component %v", currentComp)
	}
	err := child.rUnlock(remainingComp)
	if err != nil {
		// TODO: what is the right thing to do here?  Should we
		// attempt to unlock ourselves even if children failed to
		// unlock correctly?  Either doing so or not doing so seems
		// dangerous.
		return err
	}
	return nil
}

// rUnlock unlocks a path that has already been previously
// locked for writer access.  It decrements the writer count on
// the final component in the path and removes a write intention
// on all other paths.
func (tn *tagLockerNode) unlock(components []string) error {
	// If we have no paths to traverse, unlock ourselves!
	if len(components) == 0 {
		//fmt.Fprintf(os.Stderr, "XUnLock %v\n", tn.component)
		tn.tagLock.XUnlock()
		return nil
	}

	currentComp := components[0]
	remainingComp := components[1:]

	defer func() {
		//fmt.Fprintf(os.Stderr, "IXUnLock %v\n", tn.component)
		tn.tagLock.IXUnlock()
	}()

	// Unlock our children first - we want to unlock in the opposite
	// order that we acquired the locks.
	tn.mtx.RLock()
	child, ok := tn.children[currentComp]
	tn.mtx.RUnlock()
	if !ok {
		return fmt.Errorf("missing component %v", currentComp)
	}
	err := child.unlock(remainingComp)
	if err != nil {
		// TODO: what is the right thing to do here?  Should we
		// attempt to unlock ourselves even if children failed to
		// unlock correctly?  Either doing so or not doing so seems
		// dangerous.
		return err
	}
	return nil
}

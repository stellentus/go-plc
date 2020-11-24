package plc

import (
	"errors"
	"reflect"
)

type Cache struct {
	reader Reader
	cache  map[string]interface{}
}

var _ = Reader(&Cache{}) // Compiler makes sure this type is a Reader

// NewCache returns a Cache which caches the most recent value passed through it.
// Values are cached by reading them through NewCache as a Reader.
// Cached values are accessed through CacheReader.
func NewCache(reader Reader) *Cache {
	return &Cache{
		reader: reader,
		cache:  map[string]interface{}{},
	}
}

func (r *Cache) ReadTag(name string, value interface{}) error {
	err := r.reader.ReadTag(name, value)
	if err != nil {
		return err
	}

	r.cache[name] = reflect.Indirect(reflect.ValueOf(value)).Interface()
	return nil
}

// ReadCachedTag acts the same as ReadTag, but returns the cached value.
// A read of a value not in the cache will return TagNotFoundError.
func (r *Cache) ReadCachedTag(name string, value interface{}) error {
	cVal, ok := r.cache[name]
	if !ok {
		return TagNotFoundError{name}
	}

	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr {
		return errors.New("ReadCachedTag for '" + name + "' requires a reference, not a value")
	}
	vToSet := val.Elem()
	if !vToSet.CanSet() {
		return errors.New("Provided value for tag '" + name + "' cannot be set")
	}

	vToSet.Set(reflect.ValueOf(cVal))
	return nil
}

type CacheReader struct {
	cache *Cache
}

// CacheReader returns a Reader which calls ReadCachedTag.
func (r *Cache) CacheReader() CacheReader {
	return CacheReader{cache: r}
}

func (r CacheReader) ReadTag(name string, value interface{}) error {
	return r.cache.ReadCachedTag(name, value)
}

type TagNotFoundError struct {
	Name string
}

func (err TagNotFoundError) Error() string {
	return "Tag '" + err.Name + "' could not be found"
}

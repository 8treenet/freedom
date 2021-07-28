package internal

import (
	"errors"
)

// store .
type store struct {
	cache map[interface{}]interface{}
}

func newStore() *store {
	m := new(store)
	m.cache = make(map[interface{}]interface{})
	return m
}

// SetOrStore .
func (s *store) SetOrStore(key interface{}, value interface{}) (v interface{}, set bool) {
	v, set = s.cache[key]
	if set {
		set = false
		return
	}
	set = true
	s.cache[key] = value
	v = value
	return
}

// Set .
func (s *store) Set(key interface{}, value interface{}) {
	s.cache[key] = value
}

// Get .
func (s *store) Get(key interface{}, value interface{}) error {
	v, ok := s.cache[key]
	if !ok {
		return errors.New("undefined")
	}

	return ConvertAssign(value, v)
}

// Exist .
func (s *store) Exist(key interface{}) bool {
	_, ok := s.cache[key]
	return ok
}

// ToInterface .
func (s *store) ToInterface(key interface{}) interface{} {
	v, ok := s.cache[key]
	if !ok {
		return nil
	}

	return v
}

// Remove .
func (s *store) Remove(key interface{}) {
	delete(s.cache, key)
}

// RemoveAll .
func (s *store) RemoveAll() {
	s.cache = make(map[interface{}]interface{})
}

// Keys .
func (s *store) Keys() []interface{} {
	list := make([]interface{}, 0, len(s.cache))
	for k := range s.cache {
		list = append(list, k)
	}
	return list
}

// Values .
func (s *store) Values() []interface{} {
	list := make([]interface{}, 0, len(s.cache))
	for _, v := range s.cache {
		list = append(list, v)
	}
	return list
}

// ToMap .
func (s *store) ToMap() map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	for key, value := range s.cache {
		result[key] = value
	}
	return result
}

package internal

import (
	"errors"
)

// Store .
type Store struct {
	cache map[interface{}]interface{}
}

func newStore() *Store {
	m := new(Store)
	m.cache = make(map[interface{}]interface{})
	return m
}

// SetOrStore .
func (s *Store) SetOrStore(key interface{}, value interface{}) (v interface{}, set bool) {
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
func (s *Store) Set(key interface{}, value interface{}) {
	s.cache[key] = value
}

// Get .
func (s *Store) Get(key interface{}, value interface{}) error {
	v, ok := s.cache[key]
	if !ok {
		return errors.New("undefined")
	}

	return ConvertAssign(value, v)
}

// Exist .
func (s *Store) Exist(key interface{}) bool {
	_, ok := s.cache[key]
	return ok
}

// ToInterface .
func (s *Store) ToInterface(key interface{}) interface{} {
	v, ok := s.cache[key]
	if !ok {
		return nil
	}

	return v
}

// Remove .
func (s *Store) Remove(key interface{}) {
	delete(s.cache, key)
}

// RemoveAll .
func (s *Store) RemoveAll() {
	s.cache = make(map[interface{}]interface{})
}

// Keys .
func (s *Store) Keys() []interface{} {
	list := make([]interface{}, 0, len(s.cache))
	for k := range s.cache {
		list = append(list, k)
	}
	return list
}

// Values .
func (s *Store) Values() []interface{} {
	list := make([]interface{}, 0, len(s.cache))
	for _, v := range s.cache {
		list = append(list, v)
	}
	return list
}

// ToMap .
func (s *Store) ToMap() map[interface{}]interface{} {
	result := make(map[interface{}]interface{})
	for key, value := range s.cache {
		result[key] = value
	}
	return result
}

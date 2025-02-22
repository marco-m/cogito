// Package sets is a minimal implementation of a generic set data structure.

package sets

import (
	"fmt"
	"sort"

	"golang.org/x/exp/constraints"
)

// Set is a minimal set that takes only ordered types: any type that supports the
// operators < <= >= >.
type Set[T constraints.Ordered] struct {
	items map[T]struct{}
}

// New returns an empty set with capacity size. The capacity will grow and shrink as a
// stdlib map.
func New[T constraints.Ordered](size int) *Set[T] {
	return &Set[T]{items: make(map[T]struct{}, size)}
}

// From returns a set from elements.
func From[T constraints.Ordered](elements ...T) *Set[T] {
	s := New[T](len(elements))
	for _, i := range elements {
		s.items[i] = struct{}{}
	}
	return s
}

// String returns a string representation of s, ordered. This allows to simply pass a
// sets.Set as parameter to a function that expects a fmt.Stringer interface and obtain
// a comparable string.
func (s *Set[T]) String() string {
	return fmt.Sprint(s.OrderedList())
}

func (s *Set[T]) Size() int {
	return len(s.items)
}

// OrderedList returns a slice of the elements of s, ordered.
// TODO This can probably be replaced in Go 1.20 when a generics slice packages reaches
// the stdlib.
func (s *Set[T]) OrderedList() []T {
	elements := make([]T, 0, len(s.items))
	for e := range s.items {
		elements = append(elements, e)
	}
	sort.Slice(elements, func(i, j int) bool {
		return elements[i] < elements[j]
	})
	return elements
}

// Contains returns true if s contains item.
func (s *Set[T]) Contains(item T) bool {
	_, found := s.items[item]
	return found
}

// Remove deletes item from s. Returns true if the item was present.
func (s *Set[T]) Remove(item T) bool {
	if !s.Contains(item) {
		return false
	}
	delete(s.items, item)
	return true
}

// Difference returns a set containing the elements of s that are not in x.
func (s *Set[T]) Difference(x *Set[T]) *Set[T] {
	result := New[T](max(0, s.Size()-x.Size()))
	for i := range s.items {
		if !x.Contains(i) {
			result.items[i] = struct{}{}
		}
	}
	return result
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

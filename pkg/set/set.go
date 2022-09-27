package set

// Of initializes a new Set with the given values.
func Of[T comparable](vals ...T) Set[T] {
	res := make(Set[T])
	for _, val := range vals {
		res[val] = struct{}{}
	}
	return res
}

// Set implementation backed by Go maps.
type Set[T comparable] map[T]struct{}

// ToSlice converts a Set to a slice.
func (s Set[T]) ToSlice() []T {
	var (
		res = make([]T, len(s))
		i   int
	)
	for v := range s {
		res[i] = v
		i++
	}
	return res
}

// Intersection returns a Set of members common to both given sets.
func Intersection[T comparable](s1, s2 Set[T]) Set[T] {
	res := make(Set[T])
	for v := range s1 {
		if _, ok := s2[v]; ok {
			res[v] = struct{}{}
		}
	}
	return res
}

// Sub returns the Set of subtraction s1 - s2.
func Sub[T comparable](s1, s2 Set[T]) Set[T] {
	res := make(Set[T])
	for v := range s1 {
		if _, ok := s2[v]; !ok {
			res[v] = struct{}{}
		}
	}
	return res
}

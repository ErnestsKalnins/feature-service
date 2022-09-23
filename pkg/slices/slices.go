package slices

func Map[T1, T2 any](fn func(T1) T2, vals ...T1) []T2 {
	res := make([]T2, len(vals))
	for i := range vals {
		res[i] = fn(vals[i])
	}
	return res
}

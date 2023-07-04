package slices

func MapIndex[S any, T any](slice []S, fn func(int, S) T) []T {
	result := make([]T, len(slice))
	for i, item := range slice {
		result[i] = fn(i, item)
	}

	return result
}

func ToMap[S any, K comparable, V any](slice []S, keyFn func(S) K, valueFn func(S) V) map[K]V {
	result := make(map[K]V, len(slice))
	for _, item := range slice {
		result[keyFn(item)] = valueFn(item)
	}

	return result
}

func NoChangeFunc[S any]() func(S) S {
	return func(item S) S {
		return item
	}
}

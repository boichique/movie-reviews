package slices

func MapIndex[S any, T any](slice []S, fn func(int, S) T) []T {
	result := make([]T, len(slice))
	for i, item := range slice {
		result[i] = fn(i, item)
	}

	return result
}

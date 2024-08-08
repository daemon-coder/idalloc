package util

func SliceRepeat[T any](s T, count int) (result []T) {
	result = make([]T, count)
	for i := 0; i < count; i++ {
		result[i] = s
	}
	return
}

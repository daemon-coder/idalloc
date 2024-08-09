package util

func Ptr[T any](v T) *T {
	return &v
}

func ToInterfaceSlice[T any](values []T) []interface{} {
	result := make([]interface{}, len(values))
	for i, v := range values {
		result[i] = v
	}
	return result
}

package utils

func Contains[T comparable](array []T, target T) bool {
	for _, e := range array {
		if e == target {
			return true
		}
	}
	return false
}

package utils

// contains gets a slice of strings and a string and returns whether the string is in the slice
func Contains(strings []string, element string) bool {
	for _, str := range strings {
		if str == element {
			return true
		}
	}
	return false
}

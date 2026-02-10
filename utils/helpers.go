package utils

// DerefInt32 returns the value of an int32 pointer or 0 if it's nil.
func DerefInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// DerefString returns the value of a string pointer or empty string if it's nil.
func DerefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// DerefBool returns the value of a bool pointer or false if it's nil.
func DerefBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}

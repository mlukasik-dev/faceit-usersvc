package deref

// String dereferences string pointer.
func String(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// StringAddr returns pointer to the given string.
func StringAddr(s string) *string {
	return &s
}

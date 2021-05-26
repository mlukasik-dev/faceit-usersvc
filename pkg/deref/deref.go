package deref

// String dereferences string pointer
func String(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

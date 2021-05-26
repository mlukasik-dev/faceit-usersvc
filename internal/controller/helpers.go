package controller

func contains(sl []string, s string) bool {
	for _, a := range sl {
		if a == s {
			return true
		}
	}
	return false
}

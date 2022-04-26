package apt

func InSlice(slice []string, s string) bool {
	for _, x := range slice {
		if x == s {
			return true
		}
	}

	return false
}

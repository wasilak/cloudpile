package libs

// Deduplicate returns a new slice with duplicates values removed.
func Deduplicate(s []string) []string {
	if len(s) == 0 {
		return s
	}

	result := []string{}
	seen := make(map[string]struct{})
	for _, val := range s {
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

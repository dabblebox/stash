package slice

// Subset ...
func Subset(s1, s2 []string) bool {
	set := make(map[string]bool)
	for _, v := range s2 {
		set[v] = true
	}

	for _, i := range s1 {
		if _, found := set[i]; !found {
			return false
		}
	}

	return true
}

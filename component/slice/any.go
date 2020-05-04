package slice

// AnyIn ...
func AnyIn(s1, s2 []string) bool {

	for _, i2 := range s2 {
		for _, i1 := range s1 {
			if i1 == i2 {
				return true
			}
		}
	}

	return false
}

// In ...
func In(s1 string, s2 []string) bool {
	return AnyIn([]string{s1}, s2)
}

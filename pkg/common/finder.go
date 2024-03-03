package common

func FindStrInSlice(x string, y []string) bool {
	for _, n := range y {
		if x == n {
			return true
		}
	}
	return false
}

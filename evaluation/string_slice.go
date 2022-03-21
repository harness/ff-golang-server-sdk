package evaluation

func contains(input []string, q string) bool {
	for _, val := range input {
		if val == q {
			return true
		}
	}
	return false
}

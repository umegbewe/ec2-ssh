package helpers

func StrOrDefault(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	} else {
		return *s
	}
}

func filter(slice []string, predicate func(string) bool) []string {
	var result []string
	for _, elem := range slice {
		if predicate(elem) {
			result = append(result, elem)
		}
	}
	return result
}

func contains(slice []string, elem string) bool {
	for _, v := range slice {
		if v == elem {
			return true
		}
	}
	return false
}

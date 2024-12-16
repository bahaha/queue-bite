package utils

import "strings"

func Cn(classes ...string) string {
	return strings.Join(filterEmptyClasses(classes), " ")
}

func filterEmptyClasses(classes []string) []string {
	result := make([]string, 0, len(classes))
	for _, class := range classes {
		if class != "" {
			result = append(result, class)
		}
	}
	return result
}

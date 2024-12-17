package utils

import (
	"strings"

	tw "github.com/Oudwins/tailwind-merge-go"
)

func Cn(classes ...string) string {
	return tw.Merge(strings.Join(filterEmptyClasses(classes), " "))
}

func AppendClass(class string, active bool) string {
	if active {
		return class
	}
	return ""
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

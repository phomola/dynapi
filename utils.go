package dynapi

import "unicode"

func splitCamelcasedString(s string) []string {
	var (
		comps []string
		start int
	)
	runes := []rune(s)
	for i, r := range runes {
		if i > 0 && unicode.IsUpper(r) {
			comps = append(comps, string(runes[start:i]))
			start = i
		}
	}
	comps = append(comps, string(runes[start:]))
	return comps
}

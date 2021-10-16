package dynapi

import "unicode"

func SplitCamelcasedString(s string) []string {
	var (
		comps  []string
		start  int
		locked bool
	)
	runes := []rune(s)
	for i, r := range runes {
		switch {
		case i-start == 1 && unicode.IsUpper(r):
			locked = true
		case locked && unicode.IsLower(r):
			comps = append(comps, string(runes[start:i-1]))
			start = i - 1
			locked = false
		case i > 0 && !locked && unicode.IsUpper(r):
			comps = append(comps, string(runes[start:i]))
			start = i
		}
	}
	comps = append(comps, string(runes[start:]))
	return comps
}

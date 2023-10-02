package str

import "unicode"

func Capitalized(input string) string {
	if input == "" {
		return input
	}

	r := []rune(input)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

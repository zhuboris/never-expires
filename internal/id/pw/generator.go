package pw

import "math/rand"

type charOption int

const (
	digit charOption = iota
	lower
	upper
	special

	optionsCount
)

// Generate returns password with provided length, but at least 4 (4 is all options at least once)
func Generate(length int) string {
	const minLength = int(optionsCount)

	length = max(length, minLength)
	password := make([]rune, length)
	for i := range password {
		if i < minLength {
			password[i] = applyOption(charOption(i))
		} else {
			password[i] = randChar()
		}
	}

	password = shuffle(password)
	return string(password)
}

func randChar() rune {
	optionNumber := charOption(rand.Intn(int(optionsCount)))

	return applyOption(optionNumber)
}

func shuffle(password []rune) []rune {
	rand.Shuffle(len(password), func(i, j int) {
		password[i], password[j] = password[j], password[i]
	})

	return password
}

func randDigit() rune {
	return randRune('0', '9')
}

func randLower() rune {
	return randRune('a', 'z')
}

func randUpper() rune {
	return randRune('A', 'Z')
}

func randSpecial() rune {
	var (
		allowedSpecials = [...]rune{'!', '@', '#', '$', '%', '^', '&', '*'}
		count           = int32(len(allowedSpecials))
		randIndex       = rand.Int31n(count)
	)

	return allowedSpecials[randIndex]
}

func randRune(min, max rune) rune {
	offsetFromZero := max - min
	includingLastNum := offsetFromZero + 1
	return rand.Int31n(includingLastNum) + min
}

func applyOption(optionNumber charOption) rune {
	switch optionNumber {
	case digit:
		return randDigit()
	case lower:
		return randLower()
	case upper:
		return randUpper()
	case special:
		return randSpecial()
	default:
		panic("not existing charOption")
	}
}

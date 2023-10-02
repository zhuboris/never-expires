package lang

import "golang.org/x/text/language"

const defaultLanguage = "en"

type Language string

func FromLocaleIdentifier(input string) Language {
	if input == "" {
		return defaultLanguage
	}

	tag, err := language.Parse(input)
	if err != nil {
		return defaultLanguage
	}

	baseLang, _ := tag.Base()
	return Language(baseLang.String())
}

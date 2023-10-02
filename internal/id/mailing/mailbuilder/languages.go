package mailbuilder

import "github.com/zhuboris/never-expires/internal/id/lang"

type translations map[lang.Language]string

func (l translations) requestedOrDefaultValue(language lang.Language) (string, error) {
	const defaultLocale lang.Language = "en"

	if requestedValue, ok := l[language]; ok {
		return requestedValue, nil
	}

	if defaultValue, ok := l[defaultLocale]; ok {
		return defaultValue, nil
	}

	return "", ErrMissingDefaultLocale
}

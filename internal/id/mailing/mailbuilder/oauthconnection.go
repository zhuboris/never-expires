package mailbuilder

import (
	"errors"

	"github.com/zhuboris/never-expires/internal/id/lang"
	"github.com/zhuboris/never-expires/internal/id/usr/oauth"
)

type oauthConnectionTemplateInput struct {
	Subject    string
	Header     string
	Body       string
	Annotation string
}

func (b Builder) newOauthConnectionTemplateInputTemplateInput(language lang.Language, connectionType oauth.Type) (oauthConnectionTemplateInput, error) {
	content, err := b.oAuthConnectionContext(connectionType)
	if err != nil {
		return oauthConnectionTemplateInput{}, err
	}

	input, err := b.newMessageEmailTemplateInput(content, language)
	if err != nil {
		return oauthConnectionTemplateInput{}, err
	}

	return oauthConnectionTemplateInput{
		Subject:    input.subject,
		Header:     input.header,
		Body:       input.body,
		Annotation: input.annotation,
	}, nil
}

func (b Builder) oAuthConnectionContext(connectionType oauth.Type) (messageEmailContent, error) {
	switch connectionType {
	case oauth.GoogleAccount:
		return b.localesDict.GoogleConnection, nil
	case oauth.AppleID:
		return b.localesDict.AppleConnection, nil
	default:
		return messageEmailContent{}, errors.New("requested oAuth connection type not exist")
	}
}

package oauth

type Token struct {
	AuthCode string `json:"auth_code"`
	IDToken  struct {
		TokenString string `json:"token_string"`
	} `json:"id_token"`
}

func (t Token) IsMissingIDToken() bool {
	return t.IDToken.TokenString == ""
}

func (t Token) IsMissingAuthCode() bool {
	return t.AuthCode == ""
}

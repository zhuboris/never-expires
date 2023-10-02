package oauth

type LoginResultType int

const (
	Error = iota
	Login
	Connect
	Register
)

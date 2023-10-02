package usr

type column int

const (
	username column = iota + 1
	password
	email
)

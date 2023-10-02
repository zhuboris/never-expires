package usr

type (
	serviceTableName string
	oAuthMethod      func() serviceTableName
)

const (
	appleIDsTableName  = "apple_ids"
	googleIDsTableName = "google_ids"
)

func withApple() oAuthMethod {
	return func() serviceTableName {
		return appleIDsTableName
	}
}

func withGoogle() oAuthMethod {
	return func() serviceTableName {
		return googleIDsTableName
	}
}

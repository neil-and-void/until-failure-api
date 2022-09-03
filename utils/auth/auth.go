package auth

type Credentials struct {
	Username string `json:"username"`
	Email string `json:"email"`
}

func Sign(credentials Credentials) string {
	return ""
}

func Decode(credentials Credentials) string {
	return ""
}

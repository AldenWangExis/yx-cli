package auth

type TokenStore interface {
	Save(profile, token string) error
	Load(profile string) (string, bool, error)
	Delete(profile string) error
	Backend() string
}

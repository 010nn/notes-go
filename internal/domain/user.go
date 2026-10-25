package domain

type User struct {
	Username     string
	PasswordHash []byte
	NotesSalt    []byte
}

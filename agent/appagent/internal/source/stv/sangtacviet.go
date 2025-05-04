package sangtacviet

import "github.com/zrik/agent/appagent/internal/source"

const (
	BOOK_INFO_URL_LENGTH = 8
	CHAPTER_URL_LENGTH   = 9
)

type Sangtacviet struct {
	source.Book
	username string
	password string
}

func New(username, password string) source.WebSource {
	return &Sangtacviet{
		username: username,
		password: password,
	}
}

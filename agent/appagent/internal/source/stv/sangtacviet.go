package sangtacviet

import "github.com/zrik/agent/appagent/internal/source"

type Sangtacviet struct {
	source.Book
	username string
	password string
}

func New(username, password string) source.WebSource {
	return &Sangtacviet{}
}

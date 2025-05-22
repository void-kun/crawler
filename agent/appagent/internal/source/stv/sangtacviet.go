package stv

import "github.com/zrik/agent/appagent/internal/source"

type Sangtacviet struct {
	source.Book
	username string
	password string
	origin   string
}

func New(username, password, origin string) source.WebSource {
	return &Sangtacviet{
		username: username,
		password: password,
		origin:   origin,
	}
}

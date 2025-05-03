package sangtacviet

import "github.com/zrik/agent/appagent/internal/source"

type Sangtacviet struct {
	source.Book
	mappingFile string
	username    string
	password    string
}

func New(username, password, mappingFile string) source.WebSource {
	return &Sangtacviet{
		username:    username,
		password:    password,
		mappingFile: mappingFile,
	}
}

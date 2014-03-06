package skimmer

import (
	"github.com/codegangsta/martini"
	"github.com/codegangsta/martini-contrib/sessions"
)

type History interface {
	All() []string
	Add(string)
}

type SessionHistory struct {
	size    int
	name    string
	session sessions.Session
	data    []string
}

func (history *SessionHistory) All() []string {
	if history.data == nil {
		history.load()
	}
	return history.data
}

func (history *SessionHistory) Add(name string) {
	if history.data == nil {
		history.load()
	}
	history.data = append(history.data, "")
	copy(history.data[1:], history.data)
	history.data[0] = name
	history.save()
}

func (history *SessionHistory) save() {
	size := history.size
	if size > len(history.data){
		size = len(history.data)
	}
	history.session.Set(history.name, history.data[:size])
}

func (history *SessionHistory) load() {
	sessionValue := history.session.Get(history.name)
	history.data = []string{}
	if sessionValue != nil {
		if values, ok := sessionValue.([]string); ok {
			history.data = append(history.data, values...)
		}
	}

}

func NewSessionHistoryHandler(size int, name string) martini.Handler {
	return func(c martini.Context, session sessions.Session) {
		history := &SessionHistory{size: size, name: name, session: session}
		c.MapTo(history, (*History)(nil))
	}
}

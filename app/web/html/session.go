package html

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
)

type Session struct {
	*sessions.Session
}

func getSession(ctx iris.Context) (Session, error) {
	if ctx == nil {
		return Session{}, fmt.Errorf("invalid context")
	}
	sess := sessions.Get(ctx)
	if sess == nil {
		return Session{}, fmt.Errorf("unable to get session from request context")
	}
	return Session{sess}, nil
}

func (s *Session) setAuthenticated() {
	s.Set("authenticated", true)
}

func (s *Session) setUsername(username string) {
	s.Set("username", username)
}

func (s *Session) setLogout() {
	s.Delete("username")
	s.Set("authenticated", false)
}

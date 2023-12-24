package html

import (
	"fmt"
	domain "git.jfam.app/one-way-file-send"
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

func loggedInUser(ctx iris.Context) domain.UserID {
	session, err := getSession(ctx)
	if err != nil {
		return ""
	}
	return domain.UserID(session.GetString("authenticated"))
}

func (s *Session) setAuthenticated(userID domain.UserID) {
	s.Set("authenticated", string(userID))
}

func (s *Session) setUsername(username domain.Username) {
	s.Set("username", username)
}

func (s *Session) setLogout() {
	s.Delete("username")
	s.Delete("authenticated")
}

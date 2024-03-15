package web

import (
	"fmt"
	"strings"

	"github.com/benjohns1/blinkfile"
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

const flashMessageKey = "message"

func setFlashSuccess(ctx iris.Context, msg string) {
	sess := sessions.Get(ctx)
	if sess == nil {
		return
	}
	view := MessageView{SuccessMessage: msg}

	sess.SetFlash(flashMessageKey, view)
}

func setFlashErr(ctx iris.Context, a App, err error) {
	sess := sessions.Get(ctx)
	if sess == nil {
		return
	}
	view := MessageView{ErrorView: ParseAppErr(ctx, a, err)}

	sess.SetFlash(flashMessageKey, view)
}

func flashMessageView(ctx iris.Context) MessageView {
	sess := sessions.Get(ctx)
	if sess == nil {
		return MessageView{}
	}
	data := sess.GetFlash(flashMessageKey)
	if data == nil {
		return MessageView{}
	}
	msg, _ := data.(MessageView)
	return msg
}

func loggedInUser(ctx iris.Context) blinkfile.UserID {
	session, err := getSession(ctx)
	if err != nil {
		return ""
	}
	return blinkfile.UserID(session.GetString("authenticated"))
}

func (s *Session) setAuthenticated(userID blinkfile.UserID) {
	s.Set("authenticated", string(userID))
	if userID == "_admin" {
		s.Set("permission.user_management", true)
	}
}

func (s *Session) setUsername(username blinkfile.Username) {
	s.Set("username", username)
}

func (s *Session) setLogout() {
	s.Delete("username")
	s.Delete("authenticated")
	for key := range s.GetAll() {
		if strings.HasPrefix(key, "permission.") {
			s.Delete(key)
		}
	}
}

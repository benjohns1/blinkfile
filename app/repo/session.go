package repo

import (
	"context"
	"git.jfam.app/one-way-file-send/app"
	"sync"
)

type (
	Session struct {
		sessions map[app.Token]app.Session
		mu       sync.RWMutex
	}
)

func NewSession() *Session {
	return &Session{
		sessions: map[app.Token]app.Session{},
	}
}

func (r *Session) Save(_ context.Context, session app.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.Token] = session
	return nil
}

func (r *Session) GetByToken(_ context.Context, token app.Token) (app.Session, bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	found, ok := r.sessions[token]
	return found, ok, nil
}

func (r *Session) Delete(_ context.Context, token app.Token) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, token)
	return nil
}

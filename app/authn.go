package app

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"golang.org/x/crypto/argon2"
	"strings"
)

type (
	Credentials struct {
		domain.UserID
		username            domain.Username
		encodedPasswordHash string
	}
)

const (
	passwordMinLength = 16
	saltLength        = 32
)

type argon2Params struct {
	time        uint32
	memory      uint32
	parallelism uint8
	keyLength   uint32
}

var argon2DefaultParams = argon2Params{
	time:        8,
	memory:      32 * 1024,
	parallelism: 4,
	keyLength:   64,
}

func NewCredentials(userID domain.UserID, user domain.Username, pass string) (Credentials, error) {
	if userID == "" {
		return Credentials{}, fmt.Errorf("user ID cannot be empty")
	}
	if user == "" {
		return Credentials{}, fmt.Errorf("username cannot be empty")
	}
	if len(pass) < passwordMinLength {
		return Credentials{}, fmt.Errorf("password must be at least %d characters long", passwordMinLength)
	}
	encodedHash, err := hashPassword(pass, argon2DefaultParams)
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{
		UserID:              userID,
		username:            user,
		encodedPasswordHash: encodedHash,
	}, nil
}

func (c Credentials) CredentialsMatch(username domain.Username, password string) (bool, error) {
	if !stringsAreEqual(string(c.username), string(username)) {
		return false, nil
	}
	salt, hash, p, err := decodeHash(c.encodedPasswordHash)
	if err != nil {
		return false, err
	}
	checkHash := argon2Hash(salt, []byte(password), p)
	if !bytesAreEqual(hash, checkHash) {
		return false, nil
	}
	return true, nil
}

func hashPassword(pass string, p argon2Params) (string, error) {
	salt, err := generateRandomBytes(saltLength)
	if err != nil {
		return "", err
	}
	hash := argon2Hash(salt, []byte(pass), p)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, p.memory, p.time, p.parallelism, b64Salt, b64Hash)
	return encoded, nil
}

func argon2Hash(salt, pass []byte, p argon2Params) []byte {
	return argon2.IDKey(pass, salt, p.time, p.memory, p.parallelism, p.keyLength)
}

var (
	ErrInvalidHash         = fmt.Errorf("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = fmt.Errorf("incompatible version of argon2")
)

func decodeHash(encodedHash string) (salt, hash []byte, p argon2Params, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, p, ErrInvalidHash
	}
	var versionPart, paramPart, saltPart, hashPart = vals[2], vals[3], vals[4], vals[5]

	var version int
	if _, err = fmt.Sscanf(versionPart, "v=%d", &version); err != nil {
		return nil, nil, p, err
	}
	if version != argon2.Version {
		return nil, nil, p, ErrIncompatibleVersion
	}

	if _, err = fmt.Sscanf(paramPart, "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.parallelism); err != nil {
		return nil, nil, p, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(saltPart)
	if err != nil {
		return nil, nil, p, err
	}

	hash, err = base64.RawStdEncoding.Strict().DecodeString(hashPart)
	if err != nil {
		return nil, nil, p, err
	}
	p.keyLength = uint32(len(hash))

	return salt, hash, p, nil
}

func (a *App) IsAuthenticated(ctx context.Context, token Token) (bool, error) {
	if token == "" {
		return false, Error{ErrBadRequest, fmt.Errorf("session token cannot be empty")}
	}
	session, found, err := a.cfg.SessionRepo.Get(ctx, token)
	if err != nil {
		return false, Error{ErrRepo, err}
	}
	if !found {
		return false, nil
	}
	return session.isValid(a.cfg.Now), nil
}

func (a *App) Login(ctx context.Context, username domain.Username, password string, requestData SessionRequestData) (Session, error) {
	userID, err := a.authenticate(username, password)
	if err != nil {
		return Session{}, err
	}
	session, err := a.newSession(ctx, userID, requestData)
	if err != nil {
		return Session{}, err
	}
	return session, nil
}

func (a *App) Logout(ctx context.Context, token Token) error {
	err := a.cfg.SessionRepo.Delete(ctx, token)
	if err != nil {
		return Error{ErrRepo, err}
	}
	return nil
}

func (a *App) newSession(ctx context.Context, userID domain.UserID, data SessionRequestData) (Session, error) {
	token, err := a.cfg.GenerateToken()
	if err != nil {
		return Session{}, Error{ErrInternal, err}
	}
	session := Session{
		Token:              token,
		UserID:             userID,
		LoggedIn:           a.cfg.Now(),
		Expires:            a.cfg.Now().Add(a.cfg.SessionExpiration),
		SessionRequestData: data,
	}
	err = a.cfg.SessionRepo.Save(ctx, session)
	if err != nil {
		return Session{}, Error{ErrRepo, err}
	}
	return session, nil
}

func (a *App) authenticate(username domain.Username, password string) (domain.UserID, error) {
	if username == "" {
		return "", Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: username cannot be empty"),
		}
	}
	if password == "" {
		return "", Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: password cannot be empty"),
		}
	}
	credentials, found, err := a.getCredentials(username)
	if err != nil {
		return "", Error{
			Type: ErrInternal,
			Err:  fmt.Errorf("error retrieving credentials for %q: %w", username, err),
		}
	}
	if !found {
		return "", Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: no username %q found", username),
		}
	}
	match, err := credentials.CredentialsMatch(username, password)
	if err != nil {
		return "", Error{
			Type: ErrInternal,
			Err:  fmt.Errorf("error matching credentials: %w", err),
		}
	}
	if !match {
		return "", Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: passwords do not match"),
		}
	}
	return credentials.UserID, nil
}

func (a *App) getCredentials(username domain.Username) (Credentials, bool, error) {
	creds, found := a.credentials[username]
	if !found {
		return Credentials{}, false, nil
	}
	return creds, true, nil
}

func stringsAreEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func bytesAreEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

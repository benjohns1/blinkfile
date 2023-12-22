package app

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

type (
	Credentials struct {
		username            string
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

func NewCredentials(user, pass string) (Credentials, error) {
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
		username:            user,
		encodedPasswordHash: encodedHash,
	}, nil
}

func (c Credentials) CredentialsMatch(username, password string) (bool, error) {
	if !stringsAreEqual(c.username, username) {
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

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func (a App) Login(_ context.Context, username, password string) error {
	if username == "" {
		return Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: username cannot be empty"),
		}
	}
	if password == "" {
		return Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: password cannot be empty"),
		}
	}
	credentials, found, err := a.getCredentials(username)
	if err != nil {
		return Error{
			Type: ErrInternal,
			Err:  fmt.Errorf("error retrieving credentials for %q: %w", username, err),
		}
	}
	if !found {
		return Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: no username %q found", username),
		}
	}
	match, err := credentials.CredentialsMatch(username, password)
	if err != nil {
		return Error{
			Type: ErrInternal,
			Err:  fmt.Errorf("error matching credentials: %w", err),
		}
	}
	if !match {
		return Error{
			Type: ErrAuthnFailed,
			Err:  fmt.Errorf("invalid credentials: passwords do not match"),
		}
	}
	return nil
}

func (a App) getCredentials(username string) (Credentials, bool, error) {
	// TODO: multi-user
	if stringsAreEqual(username, a.cfg.AdminCredentials.username) {
		return a.cfg.AdminCredentials, true, nil
	}
	return Credentials{}, false, nil
}

func stringsAreEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func bytesAreEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type (
	Argon2id struct {
		KeyLength   uint32
		SaltLength  uint32
		Time        uint32
		Memory      uint32
		Parallelism uint8
	}
)

var (
	Argon2idDefault = Argon2id{
		KeyLength:   64,
		SaltLength:  20,
		Time:        8,
		Memory:      32 * 1024,
		Parallelism: 4,
	}

	ErrInvalidHash         = fmt.Errorf("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = fmt.Errorf("incompatible version of argon2")
)

func (h *Argon2id) Match(encodedHash string, data []byte) (matched bool, err error) {
	salt, hash, params, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}
	checkHash := params.hash(salt, data)
	if !bytesAreEqual(hash, checkHash) {
		return false, nil
	}
	return true, nil
}

var RandRead = rand.Read

func (h *Argon2id) Hash(data []byte) (encodedHash string) {
	salt := make([]byte, h.SaltLength)
	_, err := RandRead(salt)
	if err != nil {
		panic(err)
	}
	hash := h.hash(salt, data)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, h.Memory, h.Time, h.Parallelism, b64Salt, b64Hash)
	return encoded
}

func (h *Argon2id) hash(salt, pass []byte) []byte {
	return argon2.IDKey(pass, salt, h.Time, h.Memory, h.Parallelism, h.KeyLength)
}

func decodeHash(encodedHash string) (salt, hash []byte, params Argon2id, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, params, ErrInvalidHash
	}
	var versionPart, paramPart, saltPart, hashPart = vals[2], vals[3], vals[4], vals[5]

	var version int
	if _, err = fmt.Sscanf(versionPart, "v=%d", &version); err != nil {
		return nil, nil, params, fmt.Errorf("%w: %s", ErrInvalidHash, err)
	}
	if version != argon2.Version {
		return nil, nil, params, fmt.Errorf("expected version %d, found version %d: %w", argon2.Version, version, ErrIncompatibleVersion)
	}

	if _, err = fmt.Sscanf(paramPart, "m=%d,t=%d,p=%d", &params.Memory, &params.Time, &params.Parallelism); err != nil {
		return nil, nil, params, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(saltPart)
	if err != nil {
		return nil, nil, params, err
	}
	params.SaltLength, err = intToUint32(len(salt))
	if err != nil {
		return nil, nil, params, err
	}

	hash, err = base64.RawStdEncoding.Strict().DecodeString(hashPart)
	if err != nil {
		return nil, nil, params, err
	}
	params.KeyLength, err = intToUint32(len(hash))
	if err != nil {
		return nil, nil, params, err
	}

	return salt, hash, params, nil
}

const maxUint32 = int(^uint32(0))

func intToUint32(i int) (uint32, error) {
	if i < 0 {
		return 0, fmt.Errorf("conversion error: %d is less than min uint32 0", i)
	}
	if i > maxUint32 {
		return 0, fmt.Errorf("conversion error: %d is larger than max uint32 %d", i, maxUint32)
	}
	return uint32(i), nil
}

func bytesAreEqual(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}

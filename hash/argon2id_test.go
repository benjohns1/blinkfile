package hash_test

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	"github.com/benjohns1/blinkfile/hash"
	"golang.org/x/crypto/argon2"
)

func TestArgon2id_Hash(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name            string
		argon2id        hash.Argon2id
		randRead        func(b []byte) (n int, err error)
		args            args
		wantEncodedHash string
		wantPanic       any
	}{
		{
			name: "should panic if random byte reader fails",
			randRead: func([]byte) (int, error) {
				return 0, fmt.Errorf("rand byte err")
			},
			wantPanic: fmt.Errorf("rand byte err"),
		},
		{
			name: "should panic with invalid argon2id params",
			argon2id: hash.Argon2id{
				KeyLength:   0,
				SaltLength:  0,
				Time:        0,
				Memory:      0,
				Parallelism: 0,
			},
			wantPanic: "argon2: number of rounds too small",
		},
		{
			name: "should encode default argon2id params",
			randRead: func(b []byte) (int, error) {
				for i := range b {
					b[i] = 1
				}
				return 0, nil
			},
			argon2id:        hash.Argon2idDefault,
			wantEncodedHash: "$argon2id$v=19$m=32768,t=8,p=4$AQEBAQEBAQEBAQEBAQEBAQEBAQE$WGGvzWc5w6rWlPrsV8qDLVwhOsoXN0aFGLBl5bPcB0/Jc7nGhTILM0eLTJPQI6EKJuw9K5JiW4NSlaC+qkoI6A",
		},
		{
			name: "should encode custom argon2id params",
			randRead: func(b []byte) (int, error) {
				for i := range b {
					b[i] = 1
				}
				return 0, nil
			},
			argon2id: hash.Argon2id{
				KeyLength:   32,
				SaltLength:  16,
				Time:        16,
				Memory:      64 * 1024,
				Parallelism: 8,
			},
			wantEncodedHash: "$argon2id$v=19$m=65536,t=16,p=8$AQEBAQEBAQEBAQEBAQEBAQ$FUE6KSlZphSZvM0VaaCqR8/1YfgcGrk8ue2CXgHus8c",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if !reflect.DeepEqual(r, tt.wantPanic) {
					t.Errorf("Hash() panic = %#v, wantPanic %#v", r, tt.wantPanic)
				}
			}()
			if tt.randRead != nil {
				prev := hash.RandRead
				hash.RandRead = tt.randRead
				defer func() { hash.RandRead = prev }()
			}
			gotEncodedHash := tt.argon2id.Hash(tt.args.data)
			if gotEncodedHash != tt.wantEncodedHash {
				t.Errorf("Hash() gotEncodedHash = %v, want %v", gotEncodedHash, tt.wantEncodedHash)
			}
		})
	}
}

func TestArgon2id_Match(t *testing.T) {
	type args struct {
		encodedHash string
		data        []byte
	}
	tests := []struct {
		name        string
		argon2id    hash.Argon2id
		args        args
		wantMatched bool
		wantErr     error
	}{
		{
			name: "should fail if encoded hash is empty",
			args: args{
				encodedHash: "",
			},
			wantErr: hash.ErrInvalidHash,
		},
		{
			name: "should fail if parsing the version fails",
			args: args{
				encodedHash: "$$v=not-a-number$$$",
			},
			wantErr: fmt.Errorf("%w: expected integer", hash.ErrInvalidHash),
		},
		{
			name: "should fail if the version is incompatible",
			args: args{
				encodedHash: "$$v=0$$$",
			},
			wantErr: fmt.Errorf("expected version %d, found version 0: %w", argon2.Version, hash.ErrIncompatibleVersion),
		},
		{
			name: "should fail if the memory value is invalid",
			args: args{
				encodedHash: "$$v=19$m=abc$$",
			},
			wantErr: fmt.Errorf("expected integer"),
		},
		{
			name: "should fail if the time value is invalid",
			args: args{
				encodedHash: "$$v=19$m=1024,t=abc$$",
			},
			wantErr: fmt.Errorf("expected integer"),
		},
		{
			name: "should fail if the parallelism value is invalid",
			args: args{
				encodedHash: "$$v=19$m=1024,t=8,p=abc$$",
			},
			wantErr: fmt.Errorf("expected integer"),
		},
		{
			name: "should fail if the salt value has invalid characters",
			args: args{
				encodedHash: "$$v=19$m=1024,t=8,p=4$non-base64-characters$hash",
			},
			wantErr: base64.CorruptInputError(3),
		},
		{
			name: "should fail if the hash value has invalid characters",
			args: args{
				encodedHash: "$$v=19$m=1024,t=8,p=4$salt$non-base64-characters",
			},
			wantErr: base64.CorruptInputError(3),
		},
		{
			name: "should return false if the hashed value doesn't match the given value",
			args: args{
				encodedHash: "$$v=19$m=1024,t=8,p=4$salt$hash",
				data:        []byte("doesn't match"),
			},
			wantMatched: false,
		},
		{
			name: "should return false if the hashed value doesn't match the given value",
			args: args{
				encodedHash: "$$v=19$m=1024,t=8,p=4$salt$17NlmtmpHcXp2qeGJnfSMS/CIsfX6Amyd67G44i2nvNA1zv4ZyLk7OHVcliqdrElhXSEZPCivWuG2EmjsCitiw",
				data:        []byte("password"),
			},
			wantMatched: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMatched, err := tt.argon2id.Match(tt.args.encodedHash, tt.args.data)
			if !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("Match() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotMatched != tt.wantMatched {
				t.Errorf("Match() gotMatched = %v, want %v", gotMatched, tt.wantMatched)
			}
		})
	}
}

package hash

import (
	"encoding/base64"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultParams(t *testing.T) {
	pDefault := &params{
		Argon2Variant: "argon2id",
		Argon2Version: 19,
		Memory:        64 * 1024,
		Iterations:    3,
		Parallelism:   2,
		SaltLen:       16,
		KeyLen:        32,
	}

	pCurrent := getDefaultParams()

	if !reflect.DeepEqual(pDefault, pCurrent) {
		t.Fatalf("Passwors hashing parameters changed:\nexpected:\t%v\ngot:\t\t%v\n", pDefault, pCurrent)
	}
}

func TestGenerateFromPassword(t *testing.T) {
	p := getDefaultParams()

	prefix := fmt.Sprintf(
		"$%s$v=%d$m=%d,t=%d,p=%d$",
		p.Argon2Variant, p.Argon2Version, p.Memory, p.Iterations, p.Parallelism,
	)

	password := []byte("secret^+-ä75736437Lk'de")

	hash1, err := GenerateFromPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(hash1, prefix) {
		t.Errorf("prefix mismatch: %s != %s", prefix, hash1)
	}

	hash2, err := GenerateFromPassword(password)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(hash2, prefix) {
		t.Errorf("prefix mismatch: %s != %s", prefix, hash2)
	}

	if hash1 == hash2 {
		t.Errorf("same password must have different hashes: %s", hash1)
	}
}

func TestParseHash(t *testing.T) {
	testHash := "$argon2id$v=19$m=65536,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA"
	testKey, err := base64.RawStdEncoding.DecodeString("Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA")
	if err != nil {
		t.Fatalf("base64 encoding failed: %v", err)
	}
	testSalt, err := base64.RawStdEncoding.DecodeString("72ia6kK2mcDJWl0/X2o4EQ")
	if err != nil {
		t.Fatalf("base64 encoding failed: %v", err)
	}
	expectedHashData := hashData{
		Params: &params{
			Argon2Variant: "argon2id",
			Argon2Version: 19,
			Memory:        65536,
			Iterations:    3,
			Parallelism:   2,
			SaltLen:       16,
			KeyLen:        32,
		},
		Key:  testKey,
		Salt: testSalt,
	}

	testData, err := parseHash(testHash)
	if err != nil {
		t.Fatalf("failed to parse hash from string: %v", err)
	}

	if !reflect.DeepEqual(expectedHashData, *testData) {
		t.Fatalf("Parsed parameters are not equal:\nexpected:\t%v\ngot:\t\t%v\n", expectedHashData, testData)
	}

	var tests = []struct {
		testname, h string
		want        error
	}{
		{"invalid prefix", "ref$argon2id$v=19$m=65536,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA", ErrInvalidHash},
		{"missing part", "$v=19$m=65536,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA", ErrInvalidHash},
		{"extra part", "$argon2id$v=19$m=65536,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA$", ErrInvalidHash},
		{"milformed version param", "$argon2id$v=a$m=65536,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA", ErrInvalidHash},
		{"unexpected new line", "$argon2id$v=19$m=65536\n,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA", ErrInvalidHash},
		{"forbidden base64 char", "$argon2id$v=19$m=65536,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDA\rdQ5z+0spTUEAO7tqBVvHA", ErrInvalidHash},
		{"empty memory amount", "$argon2id$v=19$m=,t=3,p=2$72ia6kK2mcDJWl0/X2o4EQ$Nu9PSnVbhaHuKb5iLb6JDAdQ5z+0spTUEAO7tqBVvHA", ErrInvalidHash},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			_, err := parseHash(tt.h)
			if !errors.Is(err, tt.want) {
				t.Errorf("hash: %s\ngot %v, want %v", tt.h, err, tt.want)
			}
		})
	}
}

func TestCompareHashAndPassword(t *testing.T) {
	password1 := []byte("secret^+-ä75736437Lk'de")
	password2 := []byte("NOTsecret^+-ä75736437Lk'de")

	hash1, err := GenerateFromPassword(password1)
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := GenerateFromPassword(password2)
	if err != nil {
		t.Fatal(err)
	}

	h1Data, err := parseHash(hash1)
	if err != nil {
		t.Fatal(err)
	}
	p := h1Data.Params

	prefix := "$%s$v=%d$m=%d,t=%d,p=%d"
	suffix := fmt.Sprintf("$%s$%s", base64.RawStdEncoding.EncodeToString(h1Data.Salt), base64.RawStdEncoding.EncodeToString(h1Data.Key))
	format := fmt.Sprintf("%s%s", prefix, suffix)

	var tests = []struct {
		testname, h string
		want        error
	}{
		{"invalid argon2 variant", fmt.Sprintf(format, "argon2", p.Argon2Version, p.Memory, p.Iterations, p.Parallelism), ErrHashingParametersMismatch},
		{"invalid argon2 version", fmt.Sprintf(format, p.Argon2Variant, p.Argon2Version-1, p.Memory, p.Iterations, p.Parallelism), ErrHashingParametersMismatch},
		{"invalid memory parameter", fmt.Sprintf(format, p.Argon2Variant, p.Argon2Version, p.Memory-1, p.Iterations, p.Parallelism), ErrHashingParametersMismatch},
		{"invalid iterations parameter", fmt.Sprintf(format, p.Argon2Variant, p.Argon2Version, p.Memory, p.Iterations-1, p.Parallelism), ErrHashingParametersMismatch},
		{"invalid parallelism parameter", fmt.Sprintf(format, p.Argon2Variant, p.Argon2Version, p.Memory, p.Iterations, p.Parallelism-1), ErrHashingParametersMismatch},
		{"invalid salt (empty)", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$$%s", p.Argon2Variant, p.Argon2Version, p.Memory, p.Iterations, p.Parallelism, base64.RawStdEncoding.EncodeToString(h1Data.Salt)), ErrHashingParametersMismatch},
		{"invalid key (empty)", fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$", p.Argon2Variant, p.Argon2Version, p.Memory, p.Iterations, p.Parallelism, base64.RawStdEncoding.EncodeToString(h1Data.Key)), ErrHashingParametersMismatch},
		{"invalid password", hash1, ErrPasswordMismatch},
		{"password valid", hash2, nil},
	}

	for _, tt := range tests {
		t.Run(tt.testname, func(t *testing.T) {
			err := CompareHashAndPassword(tt.h, password2)
			if !errors.Is(err, tt.want) {
				t.Errorf("password validation: %s\ngot %v, want %v", tt.h, err, tt.want)
			}
		})
	}
}

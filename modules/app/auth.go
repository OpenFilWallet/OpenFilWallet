package app

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
	"io"
)

type Permission string

var (
	PermRead  Permission = "read"
	PermWrite Permission = "write"
	PermSign  Permission = "sign"
	PermAdmin Permission = "admin"
)

var AllPermissions = []Permission{PermRead, PermWrite, PermSign, PermAdmin}
var SignPermissions = []Permission{PermRead, PermWrite, PermSign}

const (
	saltSize       = 8
	saltBase64Size = 12
)

var apiSecret []byte

type jwtPayload struct {
	Allow []Permission // Restrict calls to certain methods
}

func SetSecret(loginScrypt []byte) {
	apiSecret = loginScrypt
}

func AuthNew(allow []Permission) ([]byte, error) {
	if apiSecret == nil {
		return nil, errors.New("must be call SetSecret")
	}
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}

	p := jwtPayload{
		allow,
	}

	token, err := jwt.Sign(&p, jwt.NewHS256(append(apiSecret, salt...)))
	if err != nil {
		return nil, err
	}

	return append([]byte(base64.StdEncoding.EncodeToString(salt)), token...), nil
}

func AuthVerify(token string) ([]Permission, error) {
	saltBase64 := []byte(token)[:saltBase64Size]
	salt, err := base64.StdEncoding.DecodeString(string(saltBase64))
	if err != nil {
		return nil, err
	}

	tokenByte := []byte(token)[saltBase64Size:]
	var payload jwtPayload
	if _, err := jwt.Verify(tokenByte, jwt.NewHS256(append(apiSecret, salt...)), &payload); err != nil {
		return nil, fmt.Errorf("JWT Verification failed: %w", err)
	}

	return payload.Allow, nil
}

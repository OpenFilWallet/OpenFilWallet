package app

import (
	"errors"
	"fmt"
	"github.com/gbrlsnchs/jwt/v3"
)

var apiSecret *jwt.HMACSHA

type jwtPayload struct {
	Allow []string // Restrict calls to certain methods
}

func SetSecret(loginScrypt []byte) {
	apiSecret = jwt.NewHS256(loginScrypt)
}

func AuthNew(allow []string) ([]byte, error) {
	if apiSecret == nil {
		return nil, errors.New("must be call SetSecret")
	}

	p := jwtPayload{
		allow,
	}
	return jwt.Sign(&p, apiSecret)
}

func AuthVerify(token string) ([]string, error) {
	var payload jwtPayload
	if _, err := jwt.Verify([]byte(token), apiSecret, &payload); err != nil {
		return nil, fmt.Errorf("JWT Verification failed: %w", err)
	}

	return payload.Allow, nil
}

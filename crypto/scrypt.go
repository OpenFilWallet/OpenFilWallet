package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"golang.org/x/crypto/scrypt"
	"io"
)

const (
	// the salt is a random number of length 8
	saltSize = 8
	// scrypt key = salt + scrypt, length：40 = 8 + 32
	scryptKeySize = 40
)

var (
	ErrInvalidLength      = errors.New("invalid scrypt key length")
	ErrMismatchedPassword = errors.New("password does not match scrypt key")
)

func Scrypt(password string) []byte {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		panic(err)
	}

	// The recommended parameters for interactive logins as of 2017 are N=32768, r=8 and p=1.
	sk, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		panic(err)
	}

	key := append(salt, sk...)

	if len(key) != scryptKeySize {
		panic("invalid length")
	}

	return key
}

func VerifyScrypt(password string, scryptKey []byte) (bool, error) {
	if len(scryptKey) != scryptKeySize {
		return false, ErrInvalidLength
	}

	salt := scryptKey[:saltSize]

	// The recommended parameters for interactive logins as of 2017 are N=32768, r=8 and p=1.
	sk, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		panic(err)
	}

	if subtle.ConstantTimeCompare(sk, scryptKey[saltSize:]) == 1 {
		return true, nil
	}

	return false, ErrMismatchedPassword
}

package app

import (
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/console/prompt"
)

func Password(isConfirm bool) (string, error) {
	password, err := prompt.Stdin.PromptPassword("Password: ")
	if err != nil {
		return "", err
	}

	if isConfirm {
		confirm, err := prompt.Stdin.PromptPassword("Confirm Password: ")
		if err != nil {
			return "", fmt.Errorf("failed to read password confirmation: %v", err)
		}
		if password != confirm {
			return "", errors.New("passwords do not match")
		}
	}

	return password, nil
}

func Confirm(msg string) (bool, error) {
	return prompt.Stdin.PromptConfirm(msg)
}

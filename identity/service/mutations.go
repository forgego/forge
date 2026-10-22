package service

import "github.com/forgego/forge/identity/repository"

func sharedTokenTransaction(users repository.UserRepository, tokens repository.TokenRepository) bool {
	return repository.SharesBuiltInTokenTransaction(users, tokens)
}

// consumeThenWrite is the fail-closed path for custom repositories without a
// shared transaction. A failed second write can consume a token, but cannot
// leave a changed credential or verified email with a reusable token.
func consumeThenWrite(consume, write func() error) error {
	if err := consume(); err != nil {
		return err
	}
	return write()
}

package authentication

type accountState interface {
	AccountStateError() error
}

func validateAccountState(user interface{}) error {
	if account, ok := user.(accountState); ok {
		if err := account.AccountStateError(); err != nil {
			return err
		}
	}
	return nil
}

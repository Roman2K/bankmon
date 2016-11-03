package bank

import "bankmon/bank/account"

type Bank interface {
	Accounts() ([]account.Account, error)
}

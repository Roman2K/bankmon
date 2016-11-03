package bbvanetcash

import (
	"encoding/json"
	"os/exec"

	bankacc "bankmon/bank/account"
)

const libDir = "ruby-bbvanetcash"

type Bank struct {
	User, Password, CompanyCode string
}

func (b Bank) Accounts() (accs []bankacc.Account, err error) {
	res, err := b.accounts()
	if err != nil {
		return
	}

	accs = make([]bankacc.Account, len(res))
	for i, acc := range res {
		accs[i] = acc
	}

	return
}

func (b Bank) accounts() (res []account, err error) {
	cmd := exec.Command("bundle", "exec", "./accounts")
	cmd.Dir = libDir
	cmd.Env = []string{
		"BANKSCRAP_USER=" + b.User,
		"BANKSCRAP_PASSWORD=" + b.Password,
		"BANKSCRAP_COMPANY_CODE=" + b.CompanyCode,
	}

	out, err := cmd.Output()
	if err != nil {
		return
	}

	err = json.Unmarshal(out, &res)
	return
}

type account struct {
	ValIBAN             string  `json:"iban"`
	ValName             string  `json:"name"`
	ValCurrency         string  `json:"currency"`
	ValBalance          float64 `json:"balance"`
	ValAvailableBalance float64 `json:"available_balance"`
}

func (a account) IBAN() string              { return a.ValIBAN }
func (a account) Name() string              { return a.ValName }
func (a account) Currency() string          { return a.ValCurrency }
func (a account) Balance() float64          { return a.ValBalance }
func (a account) AvailableBalance() float64 { return a.ValAvailableBalance }

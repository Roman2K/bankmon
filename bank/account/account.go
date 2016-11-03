package account

import "encoding/json"

type Account interface {
	IBAN() string
	Name() string
	Currency() string
	Balance() float64
	AvailableBalance() float64
}

func Marshaler(acc Account) json.Marshaler {
	return marshaler{acc}
}

type Marshal struct {
	ValIBAN             string  `json:"iban"`
	ValName             string  `json:"name"`
	ValCurrency         string  `json:"currency"`
	ValBalance          float64 `json:"balance"`
	ValAvailableBalance float64 `json:"available_balance"`
}

func (a Marshal) IBAN() string              { return a.ValIBAN }
func (a Marshal) Name() string              { return a.ValName }
func (a Marshal) Currency() string          { return a.ValCurrency }
func (a Marshal) Balance() float64          { return a.ValBalance }
func (a Marshal) AvailableBalance() float64 { return a.ValAvailableBalance }

type marshaler struct {
	acc Account
}

func (m marshaler) MarshalJSON() ([]byte, error) {
	out := Marshal{
		ValIBAN:             m.acc.IBAN(),
		ValName:             m.acc.Name(),
		ValCurrency:         m.acc.Currency(),
		ValBalance:          m.acc.Balance(),
		ValAvailableBalance: m.acc.AvailableBalance(),
	}
	return json.Marshal(out)
}

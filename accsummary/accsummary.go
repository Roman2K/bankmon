package accsummary

import (
	"fmt"
	"strings"

	"bankmon/bank/account"
	acctrace "bankmon/bank/account/trace"
	"bankmon/slackwh"
)

type Summary []Entry

func (s Summary) SlackMessage(as, text string) slackwh.M {
	attachments := make([]slackwh.M, len(s))
	for i, entry := range s {
		attachments[i] = entry.attachment()
	}
	return slackwh.M{
		"username":    as,
		"text":        text,
		"attachments": attachments,
	}
}

type Entry struct {
	Bank string
	Accs []account.Account
}

func (e Entry) attachment() slackwh.M {
	return slackwh.M{
		"fallback":    e.fallback(),
		"author_name": e.Bank,
		"fields":      e.accFields(),
	}
}

func (e Entry) fallback() string {
	avails := []string{}
	for _, acc := range e.Accs {
		out := fmt.Sprintf("%s = %s",
			acc.Name(),
			fmtMoney(acc.AvailableBalance(), acc.Currency(), false),
		)
		avails = append(avails, out)
	}
	return fmt.Sprintf("Available in %s: %s", e.Bank, strings.Join(avails, ", "))
}

func (e Entry) accFields() (fields []slackwh.M) {
	fields = make([]slackwh.M, len(e.Accs))
	for i, acc := range e.Accs {
		fields[i] = slackwh.M{
			"title": acc.Name(),
			"value": fmtBalance(acc, acc.Currency(), false),
			"short": true,
		}
	}
	return
}

type DiffSummary struct {
	Bank string
	Diff *acctrace.Diff
}

func (s DiffSummary) SlackMessage(as string) slackwh.M {
	bigger := biggerAmount(s.Diff.Balance(), s.Diff.AvailableBalance())

	color := "good"
	if bigger < 0 {
		color = "danger"
	}

	accName := s.Diff.A.Name()
	balance := fmtBalance(s.Diff, s.Diff.A.Currency(), true)

	return slackwh.M{
		"username": as,
		"text": fmt.Sprintf("New activity: %s",
			fmtMoney(bigger, s.Diff.A.Currency(), true),
		),
		"attachments": []slackwh.M{
			{
				"color":       color,
				"fallback":    fmt.Sprintf("%s: %s", accName, balance),
				"author_name": s.Bank,
				"fields": []slackwh.M{
					{
						"title": accName,
						"value": balance,
						"short": true,
					},
				},
			},
		},
	}
}

func biggerAmount(a, b float64) float64 {
	if a < 0 || b < 0 {
		if a < b {
			return a
		}
		return b
	}
	if a > b {
		return a
	}
	return b
}

type balancer interface {
	Balance() float64
	AvailableBalance() float64
}

func fmtBalance(b balancer, currency string, signed bool) (out string) {
	bal := b.Balance()
	avail := b.AvailableBalance()
	out = fmtMoney(bal, currency, signed)
	if avail != bal {
		out += fmt.Sprintf(" (available: %s)", fmtMoney(avail, currency, signed))
	}
	return
}

func fmtMoney(amt float64, currency string, signed bool) string {
	const decimals = 2
	sign := ""
	if signed {
		sign = "+"
	}
	return fmt.Sprintf("%"+sign+".*f %s", decimals, amt, currency)
}

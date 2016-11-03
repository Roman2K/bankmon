package monitor

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"

	"bankmon/accsummary"
	"bankmon/bank"
	"bankmon/bank/account"
	acctrace "bankmon/bank/account/trace"
	"bankmon/slackwh"
	"bankmon/throttle"
	"bankmon/timeutil"
)

var limitSummary throttle.Limit

func init() {
	limitSummary = throttle.Limit{
		Job: "summary",
		Comply: func(cur, prev time.Time) bool {
			// Everyday at 8 am
			curDay := timeutil.BeginningOfDay(cur)
			prevDay := timeutil.BeginningOfDay(prev)
			return curDay.After(prevDay) && cur.Hour() >= 8
		},
	}
}

type Mon struct {
	Db         *bolt.DB
	SlackBot   string
	SlackBanks string
	Banks      []NamedBank
}

type NamedBank struct {
	Name string
	Bank bank.Bank
}

// Trace = snap + diff + log
func (m *Mon) Trace() {
	now := time.Now()
	summ := accsummary.Summary{}

	// TODO parallel
	for _, b := range m.Banks {
		blog := log.WithField("bank", b.Name)
		accs, err := m.snapBank(b, blog)
		if err != nil {
			blog.Errorf("snapBank(): %v", err)
			continue
		}
		summ = append(summ, accsummary.Entry{Bank: b.Name, Accs: accs})
	}

	if err := m.logAccSummary(summ, now); err != nil {
		log.Errorf("logAccSummary(): %v", err)
	}
}

func (m *Mon) snapBank(bank NamedBank, log *log.Entry) (accs []account.Account, err error) {
	log.Debugf("fetching accounts")
	accs, err = bank.Bank.Accounts()
	if err != nil {
		return
	}
	log.Debugf("fetched %d accounts", len(accs))

	t := acctrace.NewTracer(m.Db, accs)
	err = t.Snapshot()
	if err != nil {
		return
	}
	log.Infof("snapshot %d accounts", len(accs))

	diffs, err := t.Diff()
	if err != nil {
		return
	}
	log.Debugf("found %d diffs", len(diffs))

	for _, diff := range diffs {
		dlog := log.WithField("account", diff.A.Name())
		dlog.Debugf("logging diff to Slack")
		if err := m.logAccDiff(bank.Name, diff); err != nil {
			dlog.Errorf("logAccDiff(): %v", err)
		}
	}

	return
}

func (m *Mon) logAccSummary(summ accsummary.Summary, now time.Time) (err error) {
	throt := throttle.New(m.Db)
	run, ok, err := throt.Start(limitSummary)
	if err != nil {
		return
	}
	if !ok {
		log.Debugf("limit reached, not logging account summary to Slack")
		return
	}

	log.Debugf("logging account summary to Slack")

	text := fmt.Sprintf("Daily summary on <!date^%d^{date_short}|%s>",
		now.Unix(),
		now.Format("Jan 2, 2006"),
	)
	msg := summ.SlackMessage(m.SlackBot, text)
	err = slackwh.Post(m.SlackBanks, msg)

	if runErr := run.End(); runErr != nil {
		if err == nil {
			err = runErr
		} else {
			log.Errorf("throttle.Run.End(): %v", runErr)
		}
	}

	return
}

func (m *Mon) logAccDiff(bank string, diff *acctrace.Diff) error {
	summ := accsummary.DiffSummary{Bank: bank, Diff: diff}
	msg := summ.SlackMessage(m.SlackBot)
	return slackwh.Post(m.SlackBanks, msg)
}

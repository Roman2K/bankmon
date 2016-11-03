package config

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/boltdb/bolt"

	"bankmon/bank"
	"bankmon/bbvanetcash"
	"bankmon/monitor"
)

type Config struct {
	Slack struct {
		Bot   string
		Log   string
		Banks string
	}
	DbPath     string     `json:"db_path"`
	BanksSpecs []bankSpec `json:"banks"`

	Db    *bolt.DB            `json:"-"`
	Banks []monitor.NamedBank `json:"-"`
}

func (c *Config) Load(r io.Reader) (err error) {
	err = json.NewDecoder(r).Decode(c)
	if err != nil {
		return
	}

	// Slack
	if c.Slack.Bot == "" {
		return fmt.Errorf("missing slack.bot")
	}
	if c.Slack.Log == "" {
		return fmt.Errorf("missing slack.log")
	}
	if c.Slack.Banks == "" {
		return fmt.Errorf("missing slack.banks")
	}

	// Db
	c.Db, err = bolt.Open(c.DbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("failed to open db %s: %v", c.DbPath, err)
	}

	// Banks
	c.Banks, err = initBanks(c.BanksSpecs)
	return
}

func initBanks(specs []bankSpec) (banks []monitor.NamedBank, err error) {
	for _, spec := range specs {
		b, err := initBank(spec)
		if err != nil {
			return nil, err
		}
		banks = append(banks, *b)
	}
	return
}

var initializers = map[string]func(bankSpec) (bank.Bank, error){
	"bbva": bankBbva,
}

func initBank(spec bankSpec) (*monitor.NamedBank, error) {
	typ, ok := spec["type"]
	if !ok {
		return nil, fmt.Errorf("missing bank type in %v", spec)
	}
	init, ok := initializers[typ]
	if !ok {
		return nil, fmt.Errorf("unrecognized bank type: %q", typ)
	}
	name, ok := spec["name"]
	if !ok {
		return nil, fmt.Errorf("missing bank name in %v", spec)
	}
	bank, err := init(spec)
	if err != nil {
		return nil, err
	}
	return &monitor.NamedBank{Name: name, Bank: bank}, nil
}

func bankBbva(spec bankSpec) (b bank.Bank, err error) {
	err = spec.check("user", "password", "company_code")
	if err != nil {
		return
	}
	b = bbvanetcash.Bank{
		User:        spec["user"],
		Password:    spec["password"],
		CompanyCode: spec["company_code"],
	}
	return
}

type bankSpec map[string]string

func (bs bankSpec) check(keys ...string) error {
	missing := []string{}
	for _, key := range keys {
		if _, ok := bs[key]; !ok {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing keys %s in %v", strings.Join(missing, ", "), bs)
	}
	return nil
}

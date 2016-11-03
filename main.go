package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/robfig/cron"

	"bankmon/config"
	"bankmon/monitor"
	"bankmon/slackloghook"
	"bankmon/slackwh"
)

const (
	cronSpec = "0 0,30 * * * *"
)

var (
	cronMode = flag.Bool("cron", false, "cron mode")
	logPath  = flag.String("log", "-", "log path")
)

func main() {
	if err := start(); err != nil {
		log.Fatal(err)
	}
}

func usage(w io.Writer) {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(w, "Usage: %s [<flag> ...] <config-path>\n", exe)
	fmt.Fprintf(w, "Flags:\n")
	flag.CommandLine.SetOutput(w)
	flag.PrintDefaults()
}

func start() (err error) {
	flag.Parse()

	if flag.NArg() != 1 {
		usage(os.Stderr)
		os.Exit(2)
	}
	confPath := flag.Arg(0)

	f, err := os.Open(confPath)
	if err != nil {
		return
	}
	defer f.Close()

	conf := &config.Config{}
	err = conf.Load(f)
	if err != nil {
		return
	}
	defer conf.Db.Close()

	log.SetLevel(log.DebugLevel)
	log.AddHook(slackloghook.Hook{As: conf.Slack.Bot, URL: conf.Slack.Log})

	mon := monitor.Mon{
		Db:         conf.Db,
		SlackBot:   conf.Slack.Bot,
		SlackBanks: conf.Slack.Banks,
		Banks:      conf.Banks,
	}

	trace := func() {
		timeDo("mon.Trace()", conf.Slack.Bot, conf.Slack.Log, mon.Trace)
	}

	if *cronMode {
		log.Infof("cron mode")
		return startCron(cronSpec, trace)
	}

	trace()
	return
}

func startCron(spec string, cmd func()) (err error) {
	crontab := cron.New()
	err = crontab.AddFunc(spec, cmd)
	if err != nil {
		return
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	log.Infof("starting crontab")
	crontab.Start()
	<-ch
	log.Infof("stopping crontab")
	return
}

func timeDo(desc, slackBot, slackLog string, fn func()) {
	log.Infof("starting %s", desc)
	start := time.Now()
	fn()
	elapsed := time.Now().Sub(start)
	log.Infof("finished %s in %s", desc, elapsed)
	msg := slackwh.M{
		"username": slackBot,
		"text":     fmt.Sprintf("Run %s in %s", desc, elapsed),
	}
	if err := slackwh.Post(slackLog, msg); err != nil {
		log.Errorf("slackwh.Post(): %v", err)
	}
}

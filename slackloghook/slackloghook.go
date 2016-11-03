package slackloghook

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Sirupsen/logrus"

	"bankmon/slackwh"
)

const attachmentColor = "danger"

var levels = []logrus.Level{
	logrus.ErrorLevel,
	logrus.FatalLevel,
	logrus.PanicLevel,
}

type Hook struct {
	As  string
	URL string
}

func (h Hook) Levels() []logrus.Level {
	return levels
}

func (h Hook) Fire(entry *logrus.Entry) (err error) {
	atts, err := entryAttachments(entry)
	if err != nil {
		return
	}

	msg := slackwh.M{
		"username":    h.As,
		"text":        fmt.Sprintf("ERROR (%s): %s", entry.Level.String(), entry.Message),
		"attachments": atts,
	}

	return slackwh.Post(h.URL, msg)
}

func entryAttachments(entry *logrus.Entry) (atts []slackwh.M, err error) {
	if len(entry.Data) == 0 {
		return
	}

	json, err := json.MarshalIndent(entry.Data, "", "  ")
	if err != nil {
		return
	}

	atts = []slackwh.M{
		{
			"fallback": string(json),
			"color":    attachmentColor,
			"fields":   attachmentDataFields(entry.Data),
			"ts":       entry.Time.Unix(),
		},
	}
	return
}

func attachmentDataFields(data logrus.Fields) (fields []slackwh.M) {
	names := make([]string, 0, len(data))
	for n, _ := range data {
		names = append(names, n)
	}
	sort.Strings(names)

	fields = make([]slackwh.M, 0, len(names))
	for _, n := range names {
		fields = append(fields, slackwh.M{
			"title": n,
			"value": fmt.Sprintf("%v", data[n]),
			"short": true,
		})
	}

	return
}

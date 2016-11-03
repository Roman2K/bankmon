package slackwh

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type M map[string]interface{}

func Post(url string, message M) (err error) {
	body, err := json.Marshal(message)
	if err != nil {
		return
	}

	log.Debugf("sending %d bytes to Slack incoming webhook", len(body))

	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return
	}
	defer res.Body.Close()

	err = errFromResponse(res)
	return
}

func errFromResponse(res *http.Response) error {
	if res.StatusCode == 200 {
		return nil
	}
	return errors.New(errMsgFromResponse(res))
}

func errMsgFromResponse(res *http.Response) string {
	msg, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Errorf("while reading error message in response: %v", err)
		return res.Status
	}

	return fmt.Sprintf("%s (%s)", msg, res.Status)
}

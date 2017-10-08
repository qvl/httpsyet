// Package slackhook proviodes Helpers to send messages to Slack hooks.
// For more see https://api.slack.com/incoming-webhooks.
package slackhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

// Data is sent to Slack as JSON.
type Data struct {
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	Channel   string `json:"channel,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// Post a message to a Slack incoming webhook.
func Post(hook, text string) error {
	return PostCustom(hook, Data{Text: text}, http.Post)
}

// PostCustom posts a message to Slack while allowing to overwrite the webhook defaults and http.Post.
func PostCustom(hook string, d Data, post func(string, string, io.Reader) (*http.Response, error)) error {
	buf, err := json.Marshal(d)
	if err != nil {
		return fmt.Errorf("cannot marshal json %#v: %v", d, err)
	}

	resp, err := post(hook, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return fmt.Errorf("failed to post to Slack: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %v", err)
		}
		return fmt.Errorf("HTTP status code is not OK (%d): '%s'", resp.StatusCode, body)
	}

	return nil
}

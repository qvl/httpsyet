package slack

// Format Slack message from provided output and error strings.
func Format(output, errs string) string {
	msg := output
	if errs != "" {
		msg += "\n\nErrors:\n" + errs
	}
	return msg
}

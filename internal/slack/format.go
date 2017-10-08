package slack

import (
	"strings"
)

// Format Slack message from provided output and error strings.
func Format(output, errs string) string {
	var result string

	lines := strings.Split(output, "\n")
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		if parts := strings.Split(l, " "); len(parts) == 2 {
			result += "You can change " + parts[1] + " on page " + parts[0] + " to https.\n"
		} else {
			result += strings.TrimSpace(l) + "\n"
		}
	}

	if strings.TrimSpace(errs) != "" {
		if result != "" {
			result += "\n"
		}
		result += "Errors:\n" + errs + "\n"
	}

	return result
}

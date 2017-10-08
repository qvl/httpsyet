package slack_test

import (
	"testing"

	"qvl.io/httpsyet/internal/slack"
)

func TestFormat(t *testing.T) {
	tt := []struct{ name, out, err, result string }{
		{
			name: "empty",
		},
		{
			name: "full",
			out: `https://domain.com https://external.com
https://domain.com https://external.com/sub
https://site.com https://external.com/page`,
			err: `failed to get http://expired.com
404 https://notfound.com`,
			result: `You can change https://external.com on page https://domain.com to https.
You can change https://external.com/sub on page https://domain.com to https.
You can change https://external.com/page on page https://site.com to https.

Errors:
failed to get http://expired.com
404 https://notfound.com
`,
		},
		{
			name: "no errors",
			out: `https://domain.com https://external.com
https://site.com https://external.com/page`,
			result: `You can change https://external.com on page https://domain.com to https.
You can change https://external.com/page on page https://site.com to https.
`,
		},
		{
			name: "wrong line formats",
			out: `this-is-an-invalid-format
this is also invalid`,
			result: `this-is-an-invalid-format
this is also invalid
`,
		},
		{
			name: "empty lines",
			out: `

			`,
			err: `
			`,
			result: "",
		},
		{
			name: "errors only",
			err:  "fail",
			result: `Errors:
fail
`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if r := slack.Format(tc.out, tc.err); r != tc.result {
				t.Errorf("expected:\n'%s'\n\ngot:\n'%s'", tc.result, r)
			}
		})
	}
}

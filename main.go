// Package main is the entry point for the htttpsyet binary.
// Here is where you can find argument parsing, usage information and the actual execution.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"

	"qvl.io/httpsyet/httpsyet"
	"qvl.io/httpsyet/slackhook"
)

// Can be set in build step using -ldflags
var version string

const (
	// Printed for -help, -h or with wrong number of arguments
	usage = `Find links you can update to HTTPS

Usage: %s [flags] url...

  url  one or more URLs you like to be crawled

Sites are crawled recursively. Each http:// link is check
to see if it can be replaced with https://. If a link can be replaced
it is outputed on stdout, prefixed with the site name it has been found on.
For example:

	httpsyet https://mysite.com

Might output:
	https://mysite.com http://google.com
	https://mysite.com http://facebook.com
	https://mysite.com/contact http://facebook.com
	...

Errors are reported on stderr.


Flags:
`
	more       = "\nFor more visit https://qvl.io/httpsyet."
	slackUsage = `Slack incoming webhook. If set, results are also posted to Slack.
	See https://api.slack.com/incoming-webhooks.`
)

// Get command line arguments and start crawling
func main() {
	// Flags
	slackURL := flag.String("slack", "", slackUsage)
	depth := flag.Int("depth", -1, "")
	external := flag.Bool("external", false, "")
	versionFlag := flag.Bool("version", false, "Print binary version")

	// Parse args
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(os.Stderr, more)
	}
	flag.Parse()

	if *versionFlag {
		fmt.Printf("httpsyet %s %s %s\n", version, runtime.GOOS, runtime.GOARCH)
		os.Exit(0)
	}

	sites := flag.Args()
	if len(sites) == 0 {
		flag.Usage()
		os.Exit(1)
	}

	var output io.Writer = os.Stdout
	var slackBuffer bytes.Buffer
	if *slackURL != "" {
		output = io.MultiWriter(os.Stdout, &slackBuffer)
	}
	errs := log.New(os.Stderr, "", 0)

	httpsyet.Run(httpsyet.Crawler{
		Sites:    sites,
		Out:      output,
		Logger:   errs,
		Depth:    *depth,
		External: *external,
	})

	if *slackURL == "" {
		return
	}

	if err := slackhook.Post(*slackURL, slackBuffer.String()); err != nil {
		errs.Printf("failed posting to Slack: %v", err)
		os.Exit(1)
	}
}

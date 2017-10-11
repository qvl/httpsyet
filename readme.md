#  :satellite: httpsyet :key:

[![GoDoc](https://godoc.org/qvl.io/httpsyet?status.svg)](https://godoc.org/qvl.io/httpsyet)
[![Build Status](https://travis-ci.org/qvl/httpsyet.svg?branch=master)](https://travis-ci.org/qvl/httpsyet)
[![Go Report Card](https://goreportcard.com/badge/qvl.io/httpsyet)](https://goreportcard.com/report/qvl.io/httpsyet)


The web is moving to HTTPS, [slowly](https://jorin.me/https-for-one-month/).
In a happy future, we will have secure connections only.
Today, though, we still have to deal with HTTP.
We are getting better. Thank you [Let's Encrypt](https://letsencrypt.org/).

Now we only need to update all those `http://` links on our pages to `https://`.
Not all sites support HTTPS yet. But maybe they do tomorrow.
How do we know? - `httpsyet`.

```sh
httpsyet -slack $SLACK_HOOK https://firstsite.com https://secondsite.biz http://thirdsite.net
```

This will crawl your sites recursively and for every `http://` link,
it will try if the URL is also available via HTTPS.
A list of all URLs you can update is sent to Slack.

Set this up with your favorite job scheduler ([Cron](https://en.wikipedia.org/wiki/Cron), [sleepto](https://github.com/qvl/sleepto), ...) to run once a month.


[Find out more about the implementation](https://jorin.me/use-go-channels-to-build-a-crawler/).

## Install

- With [Go](https://golang.org/):
```
go get qvl.io/httpsyet
```

- With [Homebrew](http://brew.sh/):
```
brew install qvl/tap/httpsyet
```

- Download binary: https://github.com/qvl/httpsyet/releases


## Development

Make sure to use `gofmt` and create a [Pull Request](https://github.com/qvl/httpsyet/pulls).

### Dependencies

Use [`dep ensure -update && dep prune`](https://github.com/golang/dep) to update dependencies.


### Releasing

Push a new Git tag and [GoReleaser](https://github.com/goreleaser/releaser) will automatically create a release.


## License

[MIT](./license)

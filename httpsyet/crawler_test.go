package httpsyet_test

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"qvl.io/httpsyet/httpsyet"
)

const (
	head = `
<!DOCTYPE html>
<html>
	<head lang="en">
		<meta charset="utf-8">
		<title>Page</title>
	</head>
	<body>
`

	foot = `
	</body>
</html>
`
	basePage = `
<a href="{{ .Self }}/base">
	This is an internal link to the same page. Each page should be crawled only once.
</a>

<a href="http://{{ .TLS }}/page-a">
	This link is HTTP and when checking if it also works via HTTPS, it should be a 200.
</a>

<a href="https://{{ .TLS }}/page-b">
	This link is already HTTPS. Just checking it to see if it's still working.
</a>

<a href="http://{{ .HTTP }}/page-a">
	This link is HTTP and when checking if it also works via HTTPS, it should fail. But it should still work via HTTP.
</a>

<a href="{{ .Self }}/404">
	This is an internal link which should result in a 404
</a>

<a href="/empty-sub">
	This is an relative internal link to a page without children.
</a>

<a href="{{ .Self }}/sub">
	This is an internal link.
</a>

<a href="mailto:hi@qvl.io">
	This is a mailto link. It should be ignored.
</a>

<a href="javascript:alert('hi')">
	This is a js link. It should be ignored.
</a>

`

	subPage = `
<a href="sub/sub">
	This is a relative internal link to a page without children.
</a>

<a href="http://{{ .TLS }}/page-c">
	This link is HTTP and when checking if it also works via HTTPS, it should be a 200.
</a>

<a href="http://{{ .HTTP }}/page-b">
	This link is HTTP and when checking if it also works via HTTPS, it should fail. But it should still work via HTTP.
</a>

<a href="http://{{ .HTTP }}/404">
	This link is HTTP and when checking if it also works via HTTPS, it should fail.
</a>

<a href="https://{{ .TLS }}/404">
	This link is HTTP and when checking if it also works via HTTPS, it should fail.
</a>
`

	basic = `
<h1>Welcome to the basic page</h1>
`
)

func TestRun(t *testing.T) {
	var data interface{}

	// Helpers to serve html and see which pages have been visited.
	visited := map[string]int{}
	serve := func(name, html string) http.HandlerFunc {
		visited[name] = 0
		return func(w http.ResponseWriter, r *http.Request) {
			visited[name]++
			t.Logf("visited page: %s", name)
			tmpl, err := template.New(name).Parse(head + html + foot)
			noErr(t, err)
			err = tmpl.Execute(w, data)
			noErr(t, err)
		}
	}

	pageMux := http.NewServeMux()
	pageMux.HandleFunc("/base", serve("page/base", basePage))
	pageMux.HandleFunc("/sub", serve("page/sub", subPage))
	pageMux.HandleFunc("/empty-sub", serve("page/empty-sub", basic))
	pageMux.HandleFunc("/sub/sub", serve("page/sub/sub", basic))
	pageServer := httptest.NewServer(pageMux)
	defer pageServer.Close()

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/page-a", serve("http/page-a", basic))
	httpMux.HandleFunc("/page-b", serve("http/page-b", basic))
	httpServer := httptest.NewServer(httpMux)
	defer httpServer.Close()

	tlsMux := http.NewServeMux()
	tlsMux.HandleFunc("/base", serve("tls/base", basic))
	tlsMux.HandleFunc("/base2", serve("tls/base2", basic))
	tlsMux.HandleFunc("/page-a", serve("tls/page-a", basic))
	tlsMux.HandleFunc("/page-b", serve("tls/page-b", basic))
	tlsMux.HandleFunc("/page-c", serve("tls/page-c", basic))
	tlsServer := httptest.NewTLSServer(tlsMux)
	defer tlsServer.Close()

	data = struct{ Self, HTTP, TLS string }{
		Self: pageServer.URL,
		HTTP: strings.TrimPrefix(httpServer.URL, "http://"),
		TLS:  strings.TrimPrefix(tlsServer.URL, "https://"),
	}

	var out, errs bytes.Buffer

	err := httpsyet.Crawler{
		Out: &out,
		Log: log.New(&errs, "", 0),
		Sites: []string{
			pageServer.URL + "/base",
			tlsServer.URL + "/base",
			tlsServer.URL + "/base2",
		},
		Get: tlsServer.Client().Get,
	}.Run()

	noErr(t, err)

	expect := fmt.Sprintf(
		"404 %s/404 on page %s/base\n404 %s/404 on page %s/sub\n404 %s/404 on page %s/sub\n",
		pageServer.URL,
		pageServer.URL,
		httpServer.URL,
		pageServer.URL,
		tlsServer.URL,
		pageServer.URL,
	)
	eqLines(t, expect, errs.String(), "unexpected errors")

	expect = fmt.Sprintf(
		"%s/base %s/page-a\n%s/sub %s/page-c\n",
		pageServer.URL,
		tlsServer.URL,
		pageServer.URL,
		tlsServer.URL,
	)
	eqLines(t, expect, out.String(), "unexpected output")

	for k, v := range visited {
		if v == 0 {
			t.Errorf("failed to visit page: %s", k)
		} else if v > 1 {
			t.Errorf("expected one visit on page %s; got %d", k, v)
		}
	}
}

func TestRunSingle(t *testing.T) {
	var data interface{}

	// Helpers to serve html and see which pages have been visited.
	visited := map[string]int{}
	serve := func(name, html string) http.HandlerFunc {
		visited[name] = 0
		return func(w http.ResponseWriter, r *http.Request) {
			visited[name]++
			t.Logf("visited page: %s", name)
			tmpl, err := template.New(name).Parse(head + html + foot)
			noErr(t, err)
			err = tmpl.Execute(w, data)
			noErr(t, err)
		}
	}

	pageMux := http.NewServeMux()
	pageMux.HandleFunc("/base", serve("page/base", basePage))
	pageMux.HandleFunc("/sub", serve("page/sub", subPage))
	pageMux.HandleFunc("/empty-sub", serve("page/empty-sub", basic))
	pageMux.HandleFunc("/sub/sub", serve("page/sub/sub", basic))
	pageServer := httptest.NewServer(pageMux)
	defer pageServer.Close()

	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/page-a", serve("http/page-a", basic))
	httpMux.HandleFunc("/page-b", serve("http/page-b", basic))
	httpServer := httptest.NewServer(httpMux)
	defer httpServer.Close()

	tlsMux := http.NewServeMux()
	tlsMux.HandleFunc("/page-a", serve("tls/page-a", basic))
	tlsMux.HandleFunc("/page-b", serve("tls/page-b", basic))
	tlsMux.HandleFunc("/page-c", serve("tls/page-c", basic))
	tlsServer := httptest.NewTLSServer(tlsMux)
	defer tlsServer.Close()

	data = struct{ Self, HTTP, TLS string }{
		Self: pageServer.URL,
		HTTP: strings.TrimPrefix(httpServer.URL, "http://"),
		TLS:  strings.TrimPrefix(tlsServer.URL, "https://"),
	}

	var out, errs bytes.Buffer

	err := httpsyet.Crawler{
		Out: &out,
		Log: log.New(&errs, "", 0),
		Sites: []string{
			pageServer.URL + "/base",
		},
		Get: tlsServer.Client().Get,
	}.Run()

	noErr(t, err)

	expect := fmt.Sprintf(
		"404 %s/404 on page %s/base\n404 %s/404 on page %s/sub\n404 %s/404 on page %s/sub\n",
		pageServer.URL,
		pageServer.URL,
		httpServer.URL,
		pageServer.URL,
		tlsServer.URL,
		pageServer.URL,
	)
	eqLines(t, expect, errs.String(), "unexpected errors")

	expect = fmt.Sprintf(
		"%s/base %s/page-a\n%s/sub %s/page-c\n",
		pageServer.URL,
		tlsServer.URL,
		pageServer.URL,
		tlsServer.URL,
	)
	eqLines(t, expect, out.String(), "unexpected output")

	for k, v := range visited {
		if v == 0 {
			t.Errorf("failed to visit page: %s", k)
		} else if v > 1 {
			t.Errorf("expected one visit on page %s; got %d", k, v)
		}
	}
}
func TestConfig(t *testing.T) {
	tt := []struct {
		err string
		c   httpsyet.Crawler
	}{
		{
			err: "no sites given",
			c: httpsyet.Crawler{
				Out: ioutil.Discard,
				Log: log.New(ioutil.Discard, "", 0),
			},
		},
		{
			err: "no output writer given",
			c: httpsyet.Crawler{
				Log:   log.New(ioutil.Discard, "", 0),
				Sites: []string{"https://qvl.io"},
			},
		},
		{
			err: "no error logger given",
			c: httpsyet.Crawler{
				Out:   ioutil.Discard,
				Sites: []string{"https://qvl.io"},
			},
		},
		{
			err: "depth cannot be negative",
			c: httpsyet.Crawler{
				Out:   ioutil.Discard,
				Log:   log.New(ioutil.Discard, "", 0),
				Sites: []string{"https://qvl.io"},
				Depth: -1,
			},
		},
		{
			err: "parallel cannot be negative",
			c: httpsyet.Crawler{
				Out:      ioutil.Discard,
				Log:      log.New(ioutil.Discard, "", 0),
				Sites:    []string{"https://qvl.io"},
				Parallel: -1,
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.err, func(t *testing.T) {
			doErr(t, tc.err, tc.c.Run())
		})
	}
}

func eqLines(t *testing.T, e, a, msg string) {
	es := strings.Split(e, "\n")
	sort.Strings(es)
	as := strings.Split(a, "\n")
	sort.Strings(as)
	if strings.Join(es, "\n") != strings.Join(as, "\n") {
		t.Errorf("%s; expected:\n%s\ngot:\n%s", msg, e, a)
	}
}

func noErr(t *testing.T, err error) {
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func doErr(t *testing.T, msg string, err error) {
	if err == nil {
		t.Errorf("expected error(%s); got nil", msg)
		return
	}
	if msg != err.Error() {
		t.Errorf("expected error message to be:\n%s\ngot:\n%s", msg, err.Error())
	}
}

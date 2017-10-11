// Package httpsyet provides the configuration and execution
// for crawling a list of sites for links that can be updated to HTTPS.
package httpsyet

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

const defaultParallel = 10

// Crawler is used as configuration for Run.
// Is validated in Run().
type Crawler struct {
	Sites    []string                             // At least one URL.
	Out      io.Writer                            // Required. Writes one detected site per line.
	Log      *log.Logger                          // Required. Errors are reported here.
	Depth    int                                  // Optional. Limit depth. Set to >= 1.
	Parallel int                                  // Optional. Set how many sites to crawl in parallel.
	Delay    time.Duration                        // Optional. Set delay between crawls.
	Get      func(string) (*http.Response, error) // Optional. Defaults to http.Get.
	Verbose  bool                                 // Optional. If set, status updates are written to logger.
}

type site struct {
	URL    *url.URL
	Parent *url.URL
	Depth  int
}

// Run the cralwer.
// Can return validation errors.
// All crawling errors are reported via logger.
// Output is written to writer.
// Crawls sites recursively and reports all external links that can be changed to HTTPS.
// Also reports broken links via error logger.
func (c Crawler) Run() error {
	if err := c.validate(); err != nil {
		return err
	}
	if c.Get == nil {
		c.Get = http.Get
	}
	urls, err := toURLs(c.Sites, url.Parse)
	if err != nil {
		return err
	}

	// Collect results via channel since it is not guarantied that the output writer works concurrent
	results := make(chan string)
	defer close(results)
	go func() {
		for r := range results {
			if _, err := fmt.Fprintln(c.Out, r); err != nil {
				c.Log.Printf("failed to write output '%s': %v\n", r, err)
			}
		}
	}()

	queue, sites, wait := makeQueue()

	wait <- len(urls)

	var wg sync.WaitGroup
	for i := 0; i < parallel(c.Parallel); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Pass channels to each worker
			c.worker(sites, queue, wait, results)
		}()
	}

	for _, u := range urls {
		queue <- site{
			URL:    u,
			Parent: nil,
			Depth:  c.Depth,
		}
	}

	wg.Wait()

	return nil
}

func (c Crawler) validate() error {
	if len(c.Sites) == 0 {
		return errors.New("no sites given")
	}
	if c.Out == nil {
		return errors.New("no output writer given")
	}
	if c.Log == nil {
		return errors.New("no error logger given")
	}
	if c.Depth < 0 {
		return errors.New("depth cannot be negative")
	}
	if c.Parallel < 0 {
		return errors.New("parallel cannot be negative")
	}
	return nil
}

// Returns a list of only valid URLs.
// Invalid protocols such as mailto or javascript are ignored.
// The returned error shows all invalid URLs in one message.
func toURLs(links []string, parse func(string) (*url.URL, error)) (urls []*url.URL, err error) {
	var invalids []string
	for _, s := range links {
		u, e := parse(s)
		if e != nil {
			invalids = append(invalids, fmt.Sprintf("%s (%v)", s, e))
			continue
		}
		// Default to https
		if u.Scheme == "" {
			u.Scheme = "https"
		}
		// Ignore invalid protocols
		if u.Scheme == "http" || u.Scheme == "https" {
			urls = append(urls, u)
		}
	}
	if len(invalids) > 0 {
		err = fmt.Errorf("invalid URLs: %v", strings.Join(invalids, ", "))
	}
	return
}

func parallel(p int) int {
	if p == 0 {
		return defaultParallel
	}
	return p
}

// Track visited sites via channel to prevent conflicts
// and ensure each site is visited only once.
// All channels are closed automatically as soon as queue is empty.
func makeQueue() (chan<- site, <-chan site, chan<- int) {
	queueCount := 0
	wait := make(chan int)
	sites := make(chan site)
	queue := make(chan site)
	visited := map[string]struct{}{}

	go func() {
		for delta := range wait {
			queueCount += delta
			if queueCount == 0 {
				close(queue)
			}
		}
	}()

	go func() {
		for s := range queue {
			u := s.URL.String()
			if _, v := visited[u]; !v {
				visited[u] = struct{}{}
				sites <- s
			} else {
				wait <- -1
			}
		}
		close(sites)
		close(wait)
	}()

	return queue, sites, wait
}

func (c Crawler) worker(
	sites <-chan site,
	queue chan<- site,
	wait chan<- int,
	results chan<- string,
) {
	for s := range sites {
		if c.Verbose {
			c.Log.Printf("verbose: GET %s\n", s.URL)
		}

		links, shouldUpdate, err := crawlSite(s, c.Get)

		if err != nil {
			parent := ""
			if s.Parent != nil {
				parent = fmt.Sprintf(" on page %v", s.Parent)
			}
			c.Log.Printf("%v%s\n", err, parent)
		}

		if shouldUpdate {
			s.URL.Scheme = "http"
			results <- fmt.Sprintf("%v %v", s.Parent, s.URL.String())
		}

		urls, err := toURLs(links, s.URL.Parse)
		if err != nil {
			c.Log.Printf("page %v: %v\n", s.URL, err)
		}

		wait <- len(urls) - 1

		// Submit links to queue in goroutine to not block workers
		go queueURLs(queue, urls, s.URL, s.Depth-1)

		time.Sleep(c.Delay)
	}
}

func crawlSite(s site, get func(string) (*http.Response, error)) ([]string, bool, error) {
	u := s.URL
	isExternal := s.Parent != nil && s.URL.Host != s.Parent.Host

	// If an external link is http we try https.
	// If it fails it is ignored and we carry on normally.
	// On success we return it as a result.
	if isExternal && u.Scheme == "http" {
		u.Scheme = "https"
		r2, err := get(u.String())
		if err == nil {
			defer r2.Body.Close()
			if r2.StatusCode < 400 {
				return nil, true, nil
			}
		}
		u.Scheme = "http"
	}

	r, err := get(u.String())
	if err != nil {
		return nil, false, fmt.Errorf("failed to get %v: %v", u, err)
	}
	defer r.Body.Close()

	if r.StatusCode >= 400 {
		return nil, false, fmt.Errorf("%d %v", r.StatusCode, u)
	}

	// Stop when redirecting to external page
	if r.Request.URL.Host != u.Host {
		isExternal = true
	}

	// Stop when site is external.
	// Also stop if depth one is reached, ignored when depth is set to 0.
	if isExternal || s.Depth == 1 {
		return nil, false, err
	}

	links, err := getLinks(r.Body)
	return links, false, err
}

func getLinks(r io.Reader) ([]string, error) {
	var links []string

	doc, err := html.Parse(r)
	if err != nil {
		return links, fmt.Errorf("failed to parse HTML: %v\n", err)
	}

	var f func(n *html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					links = append(links, a.Val)
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return links, nil
}

func queueURLs(queue chan<- site, urls []*url.URL, parent *url.URL, depth int) {
	for _, u := range urls {
		queue <- site{
			URL:    u,
			Parent: parent,
			Depth:  depth,
		}
	}
}

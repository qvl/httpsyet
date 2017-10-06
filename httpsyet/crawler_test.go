package httpsyet_test

import (
	"io/ioutil"
	"log"
	"testing"

	"qvl.io/httpsyet/httpsyet"
)

func TestRun(t *testing.T) {
	t.Error("oh dear...")
}

func TestNoSites(t *testing.T) {
	err := httpsyet.Crawler{
		Out: ioutil.Discard,
		Log: log.New(ioutil.Discard, "", 0),
	}.Run()
	expect(t, "no sites given", err)
}

func TestNoOutput(t *testing.T) {
	err := httpsyet.Crawler{
		Log:   log.New(ioutil.Discard, "", 0),
		Sites: []string{"https://qvl.io"},
	}.Run()
	expect(t, "no output writer given", err)
}

func TestNoLogger(t *testing.T) {
	err := httpsyet.Crawler{
		Out:   ioutil.Discard,
		Sites: []string{"https://qvl.io"},
	}.Run()
	expect(t, "no error logger given", err)
}

func TestInvalidDepth(t *testing.T) {
	err := httpsyet.Crawler{
		Out:   ioutil.Discard,
		Log:   log.New(ioutil.Discard, "", 0),
		Sites: []string{"https://qvl.io"},
		Depth: -1,
	}.Run()
	expect(t, "depth cannot be negative", err)
}

func TestInvalidParallel(t *testing.T) {
	err := httpsyet.Crawler{
		Out:      ioutil.Discard,
		Log:      log.New(ioutil.Discard, "", 0),
		Sites:    []string{"https://qvl.io"},
		Parallel: -1,
	}.Run()
	expect(t, "parallel cannot be negative", err)
}

func expect(t *testing.T, msg string, err error) {
	if err == nil {
		t.Errorf("expected error(%s); got nil", msg)
		return
	}
	if msg != err.Error() {
		t.Errorf("expected error message to be '%s'; got '%s'", msg, err.Error())
	}
}

package slackhook_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"qvl.io/httpsyet/slackhook"
)

const text = "This is a test message :heart:"

func TestSuccess(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if c := r.Header.Get("Content-Type"); c != "application/json" {
			t.Errorf("expected content-type to be application/json; got %s", c)
		}
	}))
	defer s.Close()

	if err := slackhook.Post(s.URL, text); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestServerErr(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{ "error": "bad request" }`)); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	}))
	defer s.Close()

	expected := `HTTP status code is not OK (400): '{ "error": "bad request" }'`
	err := slackhook.Post(s.URL, text)
	if err.Error() != expected {
		t.Errorf("expected error to be:\n	%s\ngot:\n	%v", expected, err)
	}
}

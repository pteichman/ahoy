package ahoy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/julienschmidt/httprouter"
)

func TestWebfinger_handleWebfingerOK(t *testing.T) {
	env := &Env{
		PublicHost: "example.org",
		PublicURL:  "https://example.org",
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
	}

	router := httprouter.New()
	router.GET("/.well-known/webfinger", handleWebfinger(env))

	req := newWebfingerRequest(t, "user@example.org")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	checkStatusCode(t, resp, http.StatusOK)
	checkHeader(t, resp, http.Header{
		"Content-Type": []string{"application/jrd+json"},
	})

	checkBody(t, resp, "{\"subject\":\"acct:user@example.org\",\"links\":[{\"rel\":\"self\",\"type\":\"application/activity+json\",\"href\":\"https://example.org/users/user\"}]}\n")
}

func TestWebfinger_handleWebfingerWrongHost(t *testing.T) {
	env := &Env{
		PublicHost: "example.org",
		PublicURL:  "https://example.org",
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
	}

	router := httprouter.New()
	router.GET("/.well-known/webfinger", handleWebfinger(env))

	req := newWebfingerRequest(t, "user@example.com")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	checkStatusCode(t, resp, http.StatusNotFound)
	checkHeader(t, resp, http.Header{
		"Content-Type":           []string{"text/plain; charset=utf-8"},
		"X-Content-Type-Options": []string{"nosniff"},
	})

	checkBody(t, resp, "not found\n")
}

func newWebfingerRequest(t *testing.T, addr string) *http.Request {
	req := http.Request{
		Method: "GET",
		URL:    newWebfingerURL(t, addr),
	}

	return &req
}

func newWebfingerURL(t *testing.T, addr string) *url.URL {
	idx := strings.Index(addr, "@")
	if idx < 0 {
		t.Fatalf("addr doesn't look like email: %s", addr)
	}

	acct := addr[:idx]
	domain := addr[idx+1:]
	resource := url.QueryEscape(fmt.Sprintf("acct:%s@%s", acct, domain))

	furl := fmt.Sprintf("https://%s/.well-known/webfinger?resource=%s", domain, resource)

	u, err := url.Parse(furl)
	if err != nil {
		t.Fatalf("failed to parse URL: %s", err)
	}

	return u
}

func checkStatusCode(t testing.TB, resp *httptest.ResponseRecorder, want int) {
	t.Helper()

	if resp.Code != want {
		t.Fatalf("resp.Code = %d, want %d", resp.Code, want)
	}
}

func checkHeader(t testing.TB, resp *httptest.ResponseRecorder, want http.Header) {
	t.Helper()

	if diff := cmp.Diff(want, resp.Header()); diff != "" {
		t.Fatalf("Header mismatch (-want +got):\n%s", diff)
	}
}

func checkBody(t testing.TB, resp *httptest.ResponseRecorder, want string) {
	t.Helper()

	got := resp.Body.String()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Body mismatch (-want +got):\n%s", diff)
	}
}

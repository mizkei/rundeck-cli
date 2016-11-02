package rundeck

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetJobLabels(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
}

func TestDo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
}

func TestAuthWithToken(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
}

func TestAuthWithPass(t *testing.T) {
	testUser := "user"
	testPassword := "password"
	testProject := "rundeck-test"

	// refs: http://rundeck.org/2.6.4/api/index.html#password-authentication
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectURL := "/api/16/j_security_check"

		r.ParseForm()
		values := r.PostForm
		user := values.Get("j_username")
		pass := values.Get("j_password")

		if r.URL.Path != expectURL {
			t.Errorf("url not match. got:%s, expect:%s", r.URL, expectURL)
		}
		if user != testUser {
			t.Errorf("username not match. got:%s, expect:%s", user, testUser)
		}
		if pass != testPassword {
			t.Errorf("password not match. got:%s, expect:%s", pass, testPassword)
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	var w bufio.Writer
	_, err = AuthWithPass(testUser, testPassword, u.Scheme, u.Host, testProject, &w)
	if err != nil {
		t.Error(err)
	}
}

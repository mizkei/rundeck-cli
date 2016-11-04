package rundeck

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestGetJobLabels(t *testing.T) {
	testToken := "token"
	testProject := "test-rundeck"

	testRes := `[
  {
    "id": "test-id-0",
    "name": "deploy",
    "group": null,
    "project": "test-rundeck",
    "description": "deploy",
    "href": "",
    "permalink": "http://test.rundeck.in/project/test-rundeck/job/show/test-id-0"
  },
  {
    "id": "test-id-1",
    "name": "done",
    "group": null,
    "project": "test-rundeck",
    "description": "done deploy",
    "href": "",
    "permalink": "http://test.rundeck.in/project/test-rundeck/job/show/test-id-1"
  }
]`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectURL := fmt.Sprintf("/api/16/project/%s/jobs", testProject)

		if r.Method != http.MethodGet {
			t.Error("http method should be GET")
		}
		if r.URL.Path != expectURL {
			t.Errorf("url not match. got:%s, expect:%s", r.URL, expectURL)
		}

		w.Write([]byte(testRes))
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	var w bufio.Writer
	rd, err := AuthWithToken(testToken, u.Scheme, u.Host, testProject, &w)
	if err != nil {
		t.Error(err)
	}

	labels, err := rd.GetJobLabels()
	if err != nil {
		t.Log(err)
	}

	expectLabels := []string{"deploy", "done"}

	if !reflect.DeepEqual(labels, expectLabels) {
		t.Error("labels not match")
	}
}

func TestDo(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()
}

func TestAuthWithToken(t *testing.T) {
	testToken := "token"
	testProject := "rundeck-test"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	var w bufio.Writer
	rd, err := AuthWithToken(testToken, u.Scheme, u.Host, testProject, &w)
	if err != nil {
		t.Error(err)
	}

	header := rd.header
	token := header.Get("X-Rundeck-Auth-Token")

	if token != testToken {
		t.Errorf("token not match. got:%s, expect:%s", token, testToken)
	}
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

		if r.Method != http.MethodPost {
			t.Error("http method should be POST")
		}
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

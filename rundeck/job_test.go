package rundeck

import (
	"bufio"
	"bytes"
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
	testToken := "token"
	testProject := "test-rundeck"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case fmt.Sprintf("/api/16/project/%s/jobs", testProject):
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
			if r.Method != http.MethodGet {
				t.Error("http method should be GET")
			}

			w.Write([]byte(testRes))
		case "/api/16/job/test-id-0":
			testRes := `- description: 'deploy'
  executionEnabled: true
  id: test-id-0
  loglevel: INFO
  name: deploy
  scheduleEnabled: true
  sequence:
    commands:
    - exec: deploy
    keepgoing: false
    strategy: node-first
  uuid: test-uuid-0
`

			if r.Method != http.MethodGet {
				t.Error("http method should be GET")
			}

			w.Write([]byte(testRes))
		case "/api/16/job/%s/executions":
		case "/api/16/execution/%d/output":
		default:
			t.Errorf("request url is wrong. url:%s", r.URL.Path)
		}
	}))
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Error(err)
	}

	rd, err := AuthWithToken(testToken, u.Scheme, u.Host, testProject, nil)
	if err != nil {
		t.Error(err)
	}

	t.Run("errors about command", func(t *testing.T) {
		var err error
		var w bytes.Buffer
		rd.out = &w

		err = rd.Do("pppp", []string{})
		if err == nil {
			t.Error("should return error message")
		}
		if err.Error() != "command 'pppp' not found" {
			t.Errorf("error message not match. got:%s, expect:%s", err.Error(), "command 'pppp' not found")
		}

		err = rd.Do(CmdRun, []string{})
		if err == nil {
			t.Error("should return error message")
		}
		if err.Error() != "job name required" {
			t.Errorf("error message not match. got:%s, expect:%s", err.Error(), "job name required")
		}

		err = rd.Do(CmdHelp, []string{})
		if err == nil {
			t.Error("should return error message")
		}
		if err.Error() != "sub command required" {
			t.Errorf("error message not match. got:%s, expect:%s", err.Error(), "sub command required")
		}

		err = rd.Do(CmdHelp, []string{"job"})
		if err == nil {
			t.Error("should return error message")
		}
		if err.Error() != "job name required" {
			t.Errorf("error message not match. got:%s, expect:%s", err.Error(), "job name required")
		}
	})

	t.Run("help jobs", func(t *testing.T) {
		var w bytes.Buffer
		rd.out = &w
		if err := rd.Do(CmdHelp, []string{SubCmdJobs}); err != nil {
			t.Error(err)
		}

		expectOut := []byte(`available jobs:

	 deploy
		 deploy

	 done
		 done deploy
`)
		if !bytes.Equal(w.Bytes(), expectOut) {
			t.Errorf("output not match.\ngot:\n%s\nexpect:\n%s", w.String(), string(expectOut))
		}
	})

	t.Run("help job", func(t *testing.T) {
		var w bytes.Buffer
		rd.out = &w
		if err := rd.Do(CmdHelp, []string{SubCmdJob, "deploy"}); err != nil {
			t.Error(err)
		}

		expectOut := []byte(`deploy
	 deploy

	options
`)
		if !bytes.Equal(w.Bytes(), expectOut) {
			t.Errorf("output not match.\ngot:\n%s\nexpect:\n%s", w.String(), string(expectOut))
		}
	})

	t.Run("run job", func(t *testing.T) {
	})
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

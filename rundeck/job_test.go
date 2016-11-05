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
	outputAPICount := 0

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
		case "/api/16/job/test-id-0/executions":
			testRes := `{
  "id": 0,
  "href": "",
  "permalink": "http://test.rundeck.in/project/test-rundeck/execution/show/0",
  "status": "running",
  "project": "test-rundeck",
  "user": "admin",
  "date-started": {
    "unixtime": 1477980000,
    "date": "2016-11-01T15:00:00Z"
  },
  "job": {
    "id": "test-id-0",
    "averageDuration": 1000,
    "name": "deploy",
    "group": "",
    "project": "test-rundeck",
    "description": "deploy",
    "href": "",
    "permalink": "http://test.rundeck.in/project/test-rundeck/job/show/test-id-0"
  },
  "description": "touch deploy.lock [... 2 steps]",
  "argstring": null
}`

			if r.Method != http.MethodPost {
				t.Error("http method should be POST")
			}

			w.Write([]byte(testRes))
		case "/api/16/execution/0/output":
			var testRes string
			if outputAPICount == 0 {
				testRes = `{
  "id": "0",
  "offset": "0",
  "completed": false,
  "message": "Pending",
  "execCompleted": false,
  "hasFailedNodes": false,
  "execState": "running",
  "execDuration": 100,
  "entries": []
}`
			} else if outputAPICount == 1 {
				testRes = `{
  "id": "0",
  "offset": "2260",
  "completed": true,
  "execCompleted": true,
  "hasFailedNodes": false,
  "execState": "succeeded",
  "lastModified": "1478336400000",
  "execDuration": 1000,
  "percentLoaded": 100,
  "totalSize": 2260,
  "entries": [
    {
      "time": "15:00:00",
      "absolute_time": "2016-11-01T15:00:00Z",
      "log": "test-log-1",
      "level": "NORMAL",
      "user": "rundeck",
      "command": null,
      "stepctx": "2",
      "node": "test.rundeck.in"
    },
    {
      "time": "15:00:00",
      "absolute_time": "2016-11-01T15:00:00Z",
      "log": "test-log-2",
      "level": "NORMAL",
      "user": "rundeck",
      "command": null,
      "stepctx": "2",
      "node": "test.rundeck.in"
    }
  ]
}`
			}

			values := r.URL.Query()
			offset := values.Get("offset")
			lastmod := values.Get("lastmod")

			if r.Method != http.MethodGet {
				t.Error("http method should be GET")
			}

			if offset != "0" {
				t.Errorf("offset not match. got:%s, expect:%s", offset, "0")
			}
			if lastmod != "0" {
				t.Errorf("lastmod not match. got:%s, expect:%s", lastmod, "0")
			}

			w.Write([]byte(testRes))
			outputAPICount++
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
		var w bytes.Buffer
		rd.out = &w
		if err := rd.Do(CmdRun, []string{"deploy"}); err != nil {
			t.Error(err)
		}

		expectOut := []byte(`job is running (http://test.rundeck.in/project/test-rundeck/execution/show/0)
test-log-1
test-log-2
done
`)
		if !bytes.Equal(w.Bytes(), expectOut) {
			t.Errorf("output not match.\ngot:\n%s\nexpect:\n%s", w.String(), string(expectOut))
		}
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

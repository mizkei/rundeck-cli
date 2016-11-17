package rundeck

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	baseURLFmt = "%s://%s/api/16"
)

type JobOption struct {
	Name       string `yaml:"name"`
	IsRequired bool   `yaml:"required"`
	Desc       string `yaml:"description"`
}

type JobDef struct {
	Name  string      `yaml:"name"`
	Desc  string      `yaml:"description"`
	Opts  []JobOption `yaml:"options"`
	Label string      `yaml:"-"`
}

type JobDefList []JobDef

type Job struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Desc      string `json:"description"`
	Permalink string `json:"permalink"`
	Label     string `json:"-"`
}

type Jobs []Job

func (js Jobs) pick(job string) *Job {
	for _, j := range js {
		if j.Label == job {
			return &j
		}
	}
	return nil
}

type Act struct {
	ID        int    `json:"id"`
	Permalink string `json:"permalink"`
}

type Entry struct {
	Log string `json:"log"`
}

type Output struct {
	Entries      []Entry `json:"entries"`
	Offset       int     `json:"offset,string"`
	LastModified int     `json:"lastModified,string"`
	Completed    bool    `json:"completed"`
}

type Rundeck struct {
	client       *http.Client
	header       http.Header
	schema, host string
	baseURL      string
	project      string
	out          io.Writer
}

func (r *Rundeck) request(method, uri string, data url.Values) (*http.Response, error) {
	u, err := url.Parse(r.baseURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, uri)

	if method == http.MethodPost {
		req, err := http.NewRequest(http.MethodPost, u.String(), strings.NewReader(data.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header = r.header

		return r.client.Do(req)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = r.header
	req.URL.RawQuery = data.Encode()

	return r.client.Do(req)
}

func (r *Rundeck) GetJobLabels() ([]string, error) {
	jobs, err := r.getJobs()
	if err != nil {
		return nil, err
	}

	labels := make([]string, 0, len(jobs))
	for _, j := range jobs {
		labels = append(labels, j.Label)
	}

	return labels, nil
}

func (r *Rundeck) getJobs() (Jobs, error) {
	res, err := r.request(http.MethodGet, fmt.Sprintf("/project/%s/jobs", r.project), url.Values{})
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var jobs Jobs
	if err := json.NewDecoder(res.Body).Decode(&jobs); err != nil {
		return nil, err
	}

	for i := range jobs {
		jobs[i].Label = normalize(jobs[i].Name)
	}

	return jobs, nil
}

func (r *Rundeck) getJobDefinition(job string) (*JobDef, error) {
	if job == "" {
		return nil, fmt.Errorf("job required")
	}

	jobs, err := r.getJobs()
	if err != nil {
		return nil, err
	}

	jb := jobs.pick(job)
	if jb == nil {
		return nil, fmt.Errorf("job(%s) not found", job)
	}

	data := url.Values{}
	data.Set("format", "yaml")
	res, err := r.request(http.MethodGet, fmt.Sprintf("/job/%s", jb.ID), data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var jdl JobDefList
	if err := yaml.Unmarshal(b, &jdl); err != nil {
		return nil, err
	}

	if len(jdl) < 1 {
		return nil, fmt.Errorf("job definition not found")
	}

	jobDef := jdl[0]
	jobDef.Label = normalize(jobDef.Name)

	return &jobDef, nil
}

func (r *Rundeck) runJob(job Job, opts []string) (*Act, error) {
	data := url.Values{}
	data.Set("argString", strings.Join(opts, " "))
	res, err := r.request(http.MethodPost, fmt.Sprintf("/job/%s/executions", job.ID), data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var act Act
	if err := json.NewDecoder(res.Body).Decode(&act); err != nil {
		return nil, err
	}

	return &act, nil
}

func (r *Rundeck) tailActivity(act Act) error {
	offset, lastmod := 0, 0
	data := url.Values{}
	var output Output
	for {
		data.Set("offset", strconv.Itoa(offset))
		data.Set("lastmod", strconv.Itoa(lastmod))

		res, err := r.request(http.MethodGet, fmt.Sprintf("/execution/%d/output", act.ID), data)
		if err != nil {
			return err
		}
		if err := json.NewDecoder(res.Body).Decode(&output); err != nil {
			return err
		}

		for _, e := range output.Entries {
			fmt.Fprintln(r.out, e.Log)
		}

		if output.Completed {
			break
		}

		offset, lastmod = output.Offset, output.LastModified

		res.Body.Close()

		time.Sleep(1 * time.Second)
	}

	return nil
}

func (r *Rundeck) run(job string, opts []string) error {
	if job == "" {
		return fmt.Errorf("job required")
	}

	jobs, err := r.getJobs()
	if err != nil {
		return err
	}

	jb := jobs.pick(job)
	if jb == nil {
		return fmt.Errorf("job(%s) not found", job)
	}

	act, err := r.runJob(*jb, opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.out, "job is running (%s)\n", act.Permalink)
	if err := r.tailActivity(*act); err != nil {
		return err
	}
	r.out.Write([]byte("done\n"))

	return nil
}

func (r *Rundeck) displayJob(jobDef JobDef) {
	fmt.Fprintln(r.out, jobDef.Label)
	fmt.Fprintln(r.out, "\t", jobDef.Desc)
	fmt.Fprintln(r.out)
	fmt.Fprintln(r.out, "\toptions")

	for _, opt := range jobDef.Opts {
		text := "optional"
		if opt.IsRequired {
			text = "required"
		}
		fmt.Fprintf(r.out, "\t\t%s (%s)\n", opt.Name, text)
		fmt.Fprintln(r.out, "\t\t\t", opt.Desc)
	}
}

func (r *Rundeck) displayJobs(jobs []Job) {
	fmt.Fprintln(r.out, "available jobs:")

	for _, job := range jobs {
		fmt.Fprintln(r.out)
		fmt.Fprintln(r.out, "\t", job.Label)

		lines := strings.Split(job.Desc, "\n")
		for _, l := range lines {
			fmt.Fprintln(r.out, "\t\t", l)
		}
	}
}

func (r *Rundeck) Do(cmd string, args []string) error {
	switch cmd {
	case CmdRun:
		if len(args) < 1 {
			return fmt.Errorf("job name required")
		}

		job, opts := args[0], args[1:]

		return r.run(job, opts)
	case CmdHelp:
		if len(args) < 1 {
			return fmt.Errorf("sub command required")
		}

		subCmd, opts := args[0], args[1:]
		switch subCmd {
		case SubCmdJobs:
			jobs, err := r.getJobs()
			if err != nil {
				return err
			}

			r.displayJobs(jobs)
		case SubCmdJob:
			if len(opts) < 1 {
				return fmt.Errorf("job name required")
			}

			jobName := opts[0]
			jobDef, err := r.getJobDefinition(jobName)
			if err != nil {
				return err
			}

			r.displayJob(*jobDef)
		default:
			return fmt.Errorf("sub command '%s' not found", subCmd)
		}
	default:
		return fmt.Errorf("command '%s' not found", cmd)
	}

	return nil
}

func AuthWithToken(token, schema, host, project string, out io.Writer) (*Rundeck, error) {
	header := http.Header{}
	header.Set("X-Rundeck-Auth-Token", token)
	header.Set("Accept", "application/json")

	baseURL := fmt.Sprintf(baseURLFmt, schema, host)

	if out == nil {
		out = os.Stdout
	}

	return &Rundeck{
		schema:  schema,
		host:    host,
		project: project,
		baseURL: baseURL,
		client:  &http.Client{},
		header:  header,
		out:     out,
	}, nil
}

func AuthWithPass(user, pass, schema, host, project string, out io.Writer) (*Rundeck, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar}

	baseURL := fmt.Sprintf(baseURLFmt, schema, host)

	signinPath := "/j_security_check"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, signinPath)
	data := url.Values{}
	data.Set("j_username", user)
	data.Set("j_password", pass)

	res, err := client.PostForm(u.String(), data)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	header := http.Header{}
	header.Set("Accept", "application/json")

	if out == nil {
		out = os.Stdout
	}

	return &Rundeck{
		schema:  schema,
		host:    host,
		project: project,
		baseURL: baseURL,
		client:  client,
		header:  header,
		out:     out,
	}, nil
}

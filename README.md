# rundeck-cli

[![Build Status](https://travis-ci.org/mizkei/rundeck-cli.svg?branch=master)](https://travis-ci.org/mizkei/rundeck-cli)

# Config

default config path is `$HOME/.config/rundeck-cli/conf.json`  
or  
`rundeck-cli -conf=conf.json`

```json
{
  "schema":  "write schema(http or https)",
  "host":    "write host name",
  "project": "write project name",
  "token":   "write token. if token is empty, authorize with password"
}
```

# Usage

## prompt mode

### commands

- run $job-name
- help {job, jobs} $job-name

sample
```
> rundeck-cli
username: rundeck
password:
rundeck>
run help
rundeck> help job
backup                 restore                system-recover
rundeck> help job
```

## command line arguments

sample
```
> rundeck-cli help jobs 
> rundeck-cli run backup
```

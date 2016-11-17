package main

import (
	"regexp"
	"strings"

	"github.com/mizkei/rundeck-cli/rundeck"
)

var (
	re  = regexp.MustCompile(` +`)
	re2 = regexp.MustCompile(`^ +`)
)

func listHasPrefix(pre string, ls []string) []string {
	list := make([]string, 0, len(ls))

	for _, s := range ls {
		if strings.HasPrefix(s, pre) {
			list = append(list, s)
		}
	}

	return list
}

type completer struct {
	cmds    []string
	subCmds []string
	jobs    []string
}

func (c *completer) completeCmd(line string, pos int) (string, []string, string) {
	pre, ls := line[:pos], line[pos:]
	pre = re2.ReplaceAllString(re.ReplaceAllString(pre, " "), "")

	newPre := pre + " "
	var list []string

	switch ss := strings.Split(pre, " "); len(ss) {
	case 1:
		target := ss[0]
		newPre = ""
		if target == "" {
			list = c.cmds
			break
		}
		list = listHasPrefix(target, c.cmds)
	case 2:
		target := ss[1]
		newPre = ss[0] + " "
		if ss[0] == rundeck.CmdRun {
			list = listHasPrefix(target, c.jobs)
			break
		}
		list = listHasPrefix(target, c.subCmds)
	case 3:
		target := ss[2]
		newPre = strings.Join(ss[:2], " ") + " "
		if ss[1] == rundeck.SubCmdJob {
			list = listHasPrefix(target, c.jobs)
		}
	}

	if len(list) == 1 {
		list[0] += " "
	}

	return newPre, list, ls
}

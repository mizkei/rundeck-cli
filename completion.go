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

	ss := strings.Split(pre, " ")

	switch len(ss) {
	case 1:
		target := ss[0]
		if target == "" {
			return "", c.cmds, ls
		}
		return "", listHasPrefix(target, c.cmds), ls
	case 2:
		target := ss[1]
		if ss[0] == rundeck.CmdRun {
			return ss[0] + " ", listHasPrefix(target, c.jobs), ls
		}

		return ss[0] + " ", listHasPrefix(target, c.subCmds), ls
	case 3:
		pre := strings.Join(ss[:2], " ")
		target := ss[2]
		if ss[1] == rundeck.SubCmdJob {
			return pre + " ", listHasPrefix(target, c.jobs), ls
		}

		return pre + " ", nil, ls
	default:
		return pre + " ", nil, ls
	}
}

package rundeck

import (
	"testing"
)

func TestCmds(t *testing.T) {
	expectCmds := []string{"run", "help"}

	cmds := Cmds()

	if len(cmds) != len(expectCmds) {
		t.Errorf("cmds length not match. got:%d, expect:%d", len(cmds), len(expectCmds))
	}

	for i := range cmds {
		if cmds[i] != expectCmds[i] {
			t.Errorf("cmd not match. got:%s, expect:%s", cmds[i], expectCmds[i])
		}
	}
}

func TestSubCmds(t *testing.T) {
	expectSubCmds := []string{"job", "jobs"}

	subCmds := SubCmds()

	if len(subCmds) != len(expectSubCmds) {
		t.Errorf("subCmds length not match. got:%d, expect:%d", len(subCmds), len(expectSubCmds))
	}

	for i := range subCmds {
		if subCmds[i] != expectSubCmds[i] {
			t.Errorf("cmd not match. got:%s, expect:%s", subCmds[i], expectSubCmds[i])
		}
	}
}

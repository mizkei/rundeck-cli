package rundeck

const (
	cmdRun  = "run"
	cmdHelp = "help"
)

const (
	subCmdJob  = "job"
	subCmdJobs = "jobs"
)

func Cmds() []string {
	return []string{cmdRun, cmdHelp}
}

func SubCmds() []string {
	return []string{subCmdJob, subCmdJobs}
}

package rundeck

const (
	CmdRun  = "run"
	CmdHelp = "help"
)

const (
	SubCmdJob  = "job"
	SubCmdJobs = "jobs"
)

func Cmds() []string {
	return []string{CmdRun, CmdHelp}
}

func SubCmds() []string {
	return []string{SubCmdJob, SubCmdJobs}
}

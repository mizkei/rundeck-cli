package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mizkei/rundeck-cli/rundeck"
	"github.com/peterh/liner"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	if !terminal.IsTerminal(0) {
		fmt.Println("no support")
		return
	}

	var confPath string
	flag.StringVar(&confPath, "conf", "$HOME/.config/go-rundeck-cli/conf.json", "config path")
	flag.Parse()

	conf, err := loadConf(confPath)
	if err != nil {
		panic(err)
	}

	line := liner.NewLiner()
	defer line.Close()
	line.SetCtrlCAborts(true)

	var rd *rundeck.Rundeck
	if conf.Token == "" {
		username, err := line.Prompt("username: ")
		if err != nil {
			panic(err)
		}
		pass, err := line.PasswordPrompt("password: ")
		if err != nil {
			panic(err)
		}

		rd, err = rundeck.AuthWithPass(username, pass, conf.Host, conf.Project, os.Stdout)
		if err != nil {
			panic(err)
		}
	} else {
		var err error
		rd, err = rundeck.AuthWithToken(conf.Token, conf.Host, conf.Project, os.Stdout)
		if err != nil {
			panic(err)
		}
	}

	for {
		l, err := line.Prompt("rundeck> ")
		if err != nil {
			fmt.Println(err)
			return
		}

		l = strings.TrimSpace(l)
		strs := strings.Split(l, " ")
		if l == "" || len(strs) == 0 {
			continue
		}

		cmd, args := strs[0], strs[1:]

		if cmd == "exit" {
			break
		}

		if err := rd.Do(cmd, args); err != nil {
			fmt.Println(err)
		}

		line.AppendHistory(l)
	}
}

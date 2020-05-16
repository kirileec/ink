package command

import (
	"bufio"
	"github.com/linxlib/logs"
	"ink/article"
	"os/exec"
	"runtime"
)

func Publish() {
	Build()

	command := article.Global.Build.Publish
	// Prepare exec command
	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "cmd"
		flag = "/C"
	} else {
		shell = "/bin/sh"
		flag = "-c"
	}
	cmd := exec.Command(shell, flag, command)
	cmd.Dir = article.RootPath
	// Start print stdout and stderr of process
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	out := bufio.NewScanner(stdout)
	err := bufio.NewScanner(stderr)
	// Print stdout
	go func() {
		for out.Scan() {
			logs.Info(out.Text())
		}
	}()
	// Print stdin
	go func() {
		for err.Scan() {
			logs.Info(err.Text())
		}
	}()
	// Exec command
	cmd.Run()
}

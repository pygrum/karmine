package kmdline

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"

	"github.com/chzyer/readline"
	"github.com/pygrum/karmine/config"
)

type kmdline struct {
	PS1        string
	ConfigPath string
	Reader     *readline.Instance
}

func Kmdline(prompt string) (*kmdline, error) {
	kl := &kmdline{}
	var err error
	kl.Reader, err = readline.NewEx(&readline.Config{
		Prompt:          "\033[31m" + prompt + "\033[0m ",
		HistoryFile:     "/tmp/kmdline.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold: true,
	})
	if err != nil {
		return nil, err
	}
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	kl.PS1 = prompt
	kl.ConfigPath = dirname + "/.konfig"
	return kl, nil
}

func (kl *kmdline) Read() error {
	l := kl.Reader
	defer l.Close()
	l.CaptureExitSignal()
	binPath, err := config.GetBinPath()
	if err != nil {
		return err
	}
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt || len(line) == 0 {
			continue
		} else if err == io.EOF {
			break
		}
		// put quoted strings into one field https://stackoverflow.com/questions/47489745/splitting-a-string-at-space-except-inside-quotation-marks
		r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)'`)
		a := r.FindAllString(line, -1)
		var tokens []string
		for _, t := range a {
			if (t[0] == '"' || t[0] == '\'') && (t[len(t)-1] == '"' || t[len(t)-1] == '\'') { // remove leftover quotes
				t = t[1:]
				t = t[:len(t)-1]
			}
			tokens = append(tokens, t)
		}
		if tokens[0] == "exit" {
			if err = os.Remove("/tmp/karmine.tmp"); err != nil {
				return err
			}
			os.Exit(0)
		}
		if tokens[0] == "clear" {
			fmt.Print("\033[H\033[2J")
			continue
		}
		c := binPath + "/" + tokens[0]
		var cout, cerr bytes.Buffer

		cmd := exec.Command(c, tokens[1:]...)
		cmd.Stdout = &cout
		cmd.Stderr = &cerr
		if err = cmd.Run(); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println(tokens[0]+":", "command not found")
			}
			if len(cerr.String()) != 0 {
				fmt.Println(cerr.String())
			}
			continue
		}
		if len(cerr.String()) != 0 {
			fmt.Println(cerr.String())
		}
		if len(cout.String()) != 0 {
			fmt.Println(cout.String())
		}
	}
	return nil
}

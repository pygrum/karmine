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
	Prompt     string
	ConfigPath string
	Reader     *readline.Instance
}

var cmdMap = make(map[string]string)

func Kmdline(prompt string) (*kmdline, error) {
	kl := &kmdline{}
	var err error
	cwd, _ := os.Getwd()
	kl.Prompt = prompt
	kl.Reader, err = readline.NewEx(&readline.Config{
		Prompt:          "\033[31m" + cwd + prompt + "\033[0m",
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
	cmdMap["new"] = binPath + "/new"
	cmdMap["profiles"] = binPath + "/profiles"
	cmdMap["stage"] = binPath + "/stage"
	cmdMap["deploy"] = binPath + "/deploy"

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
		switch tokens[0] {
		case "exit":
			if err = os.Remove("/tmp/karmine.tmp"); err != nil {
				return err
			}
			os.Exit(0)
		case "clear":
			fmt.Print("\033[H\033[2J")
			continue
		case "cd":
			if len(tokens) == 2 {
				os.Chdir(tokens[1])
				cwd, _ := os.Getwd()
				kl.Reader.SetPrompt("\033[31m" + cwd + kl.Prompt + "\033[0m")
			}
			continue
		case "help":
			fmt.Println("================")
			fmt.Println("custom commands")
			fmt.Println("================")
			for c := range cmdMap {
				fmt.Println(c)
			}
			fmt.Println("================")
			fmt.Println("run any with '--help' to view usage")
			fmt.Println()
			continue
		}
		c, ok := cmdMap[tokens[0]]
		if !ok {
			c = tokens[0]
		}
		cmd := exec.Command(c, tokens[1:]...)
		var buf bytes.Buffer
		mw := io.MultiWriter(os.Stdout, &buf)

		cmd.Stdout = mw
		cmd.Stderr = mw
		if err = cmd.Run(); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				fmt.Println(tokens[0]+":", "command not found")
			}
			continue
		}
	}
	return nil
}

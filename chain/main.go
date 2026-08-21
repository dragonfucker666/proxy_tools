package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type SubProgram struct {
	Number     int
	Name       string
	Args       []string
	InputPath  string
	OutputPath string
}

func main() {
	storageDir, groups := parseArgs()

	err := os.MkdirAll(storageDir, 0755)
	if err != nil {
		fmt.Println("chain: could not create storage directory:", err)
		os.Exit(1)
	}

	removeSocketFiles(storageDir)

	subPrograms := buildSubPrograms(storageDir, groups)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	cmds := make([]*exec.Cmd, len(subPrograms))

	aborted := false
	for i := len(subPrograms) - 1; i >= 0; i-- {
		sub := subPrograms[i]

		cmd, err := startSubProgram(sub)
		if err != nil {
			fmt.Println("chain: could not start", sub.Name, ":", err)
			aborted = true
			break
		}
		cmds[i] = cmd

		if sub.InputPath != "" {
			ok := waitForFile(sub.InputPath, signalChan)
			if !ok {
				aborted = true
				break
			}
		}
	}

	if aborted {
		killAll(cmds)
		os.Exit(1)
	}

	exitChan := make(chan int)
	for i, cmd := range cmds {
		number := i + 1
		go func(c *exec.Cmd, n int) {
			c.Wait()
			exitChan <- n
		}(cmd, number)
	}

	select {
	case n := <-exitChan:
		fmt.Println("chain: subprogram", n, "exited, stopping everything")
	case <-signalChan:
		fmt.Println("chain: got a stop signal, stopping everything")
	}

	killAll(cmds)
}

func parseArgs() (string, [][]string) {
	storage := flag.String("storage", ".", "directory for socket files")
	flag.Parse()

	rest := flag.Args()
	if len(rest) == 0 {
		fmt.Println("Usage: chain --storage ./path SEP prog1 arg SEP prog2 arg1 arg2 SEP prog3")
		os.Exit(1)
	}

	separator := rest[0]
	groups := splitGroups(rest[1:], separator)

	if len(groups) == 0 {
		fmt.Println("chain: no subprograms given")
		os.Exit(1)
	}

	return *storage, groups
}

func splitGroups(tokens []string, separator string) [][]string {
	var groups [][]string
	var current []string

	for _, token := range tokens {
		if token == separator {
			if len(current) > 0 {
				groups = append(groups, current)
				current = nil
			}
			continue
		}
		current = append(current, token)
	}

	if len(current) > 0 {
		groups = append(groups, current)
	}

	return groups
}

func removeSocketFiles(storageDir string) {
	entries, err := os.ReadDir(storageDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".socket") {
			os.Remove(filepath.Join(storageDir, entry.Name()))
		}
	}
}

func socketPath(storageDir string, programName string, number int) string {
	baseName := filepath.Base(programName)
	fileName := baseName + "_" + strconv.Itoa(number) + ".socket"
	return filepath.Join(storageDir, fileName)
}

func buildSubPrograms(storageDir string, groups [][]string) []*SubProgram {
	total := len(groups)
	subPrograms := make([]*SubProgram, total)

	for i, group := range groups {
		subPrograms[i] = &SubProgram{
			Number: i + 1,
			Name:   group[0],
			Args:   group[1:],
		}
	}

	for i, sub := range subPrograms {
		if sub.Number > 1 {
			sub.InputPath = socketPath(storageDir, sub.Name, sub.Number)
		}
		if sub.Number < total {
			nextSub := subPrograms[i+1]
			sub.OutputPath = socketPath(storageDir, nextSub.Name, nextSub.Number)
		}
	}

	return subPrograms
}

func startSubProgram(sub *SubProgram) (*exec.Cmd, error) {
	cmd := exec.Command(sub.Name, sub.Args...)
	cmd.Env = os.Environ()

	if sub.InputPath != "" {
		cmd.Env = append(cmd.Env, "INPUT="+sub.InputPath)
	}
	if sub.OutputPath != "" {
		cmd.Env = append(cmd.Env, "OUTPUT="+sub.OutputPath)
	}

	prefix := filepath.Base(sub.Name) + "_" + strconv.Itoa(sub.Number) + ": "

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	go copyWithPrefix(os.Stdout, stdoutPipe, prefix)
	go copyWithPrefix(os.Stderr, stderrPipe, prefix)

	return cmd, nil
}

func copyWithPrefix(destination io.Writer, source io.Reader, prefix string) {
	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		fmt.Fprintln(destination, prefix+scanner.Text())
	}
}

func waitForFile(path string, stop <-chan os.Signal) bool {
	for {
		_, err := os.Stat(path)
		if err == nil {
			return true
		}
		select {
		case <-stop:
			return false
		case <-time.After(50 * time.Millisecond):
		}
	}
}

func killAll(cmds []*exec.Cmd) {
	for _, cmd := range cmds {
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
	}
}

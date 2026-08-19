package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type subprogram struct {
	name string
	args []string
}

func main() {
	storage, subprograms, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "chain:", err)
		os.Exit(1)
	}

	if err := deleteSocketFiles(storage); err != nil {
		fmt.Fprintln(os.Stderr, "chain:", err)
		os.Exit(1)
	}

	sockets := socketPaths(storage, subprograms)
	runChain(subprograms, sockets)
}

func parseArgs(args []string) (string, []subprogram, error) {
	storage := "."
	if len(args) > 0 && args[0] == "--storage" {
		if len(args) < 2 {
			return "", nil, fmt.Errorf("--storage needs a value")
		}
		storage = args[1]
		args = args[2:]
	}

	if len(args) == 0 {
		return "", nil, fmt.Errorf("missing separator argument")
	}
	separator := args[0]
	args = args[1:]

	subprograms := splitIntoSubprograms(args, separator)
	if len(subprograms) == 0 {
		return "", nil, fmt.Errorf("no subprograms given")
	}
	return storage, subprograms, nil
}

func splitIntoSubprograms(args []string, separator string) []subprogram {
	var result []subprogram
	var current []string

	flush := func() {
		if len(current) > 0 {
			result = append(result, subprogram{name: current[0], args: current[1:]})
			current = nil
		}
	}

	for _, a := range args {
		if a == separator {
			flush()
			continue
		}
		current = append(current, a)
	}
	flush()

	return result
}

func deleteSocketFiles(storage string) error {
	matches, err := filepath.Glob(filepath.Join(storage, "*.socket"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// sockets[1] is left empty on purpose: the first subprogram has no input file.
func socketPaths(storage string, subprograms []subprogram) []string {
	sockets := make([]string, len(subprograms)+1)
	for i, prog := range subprograms {
		number := i + 1
		if number == 1 {
			continue
		}
		name := filepath.Base(prog.name)
		fileName := fmt.Sprintf("%s_%d.socket", name, number)
		sockets[number] = filepath.Join(storage, fileName)
	}
	return sockets
}

func buildArgs(number int, total int, sockets []string, userArgs []string) []string {
	var args []string
	if number > 1 {
		args = append(args, "--input", sockets[number])
	}
	if number < total {
		args = append(args, "--output", sockets[number+1])
	}
	args = append(args, userArgs...)
	return args
}

func runChain(subprograms []subprogram, sockets []string) {
	total := len(subprograms)
	commands := make([]*exec.Cmd, total+1)

	stop := make(chan struct{})
	var stopOnce sync.Once
	triggerStop := func() {
		stopOnce.Do(func() { close(stop) })
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		triggerStop()
	}()

	var running sync.WaitGroup

	for number := total; number >= 1; number-- {
		prog := subprograms[number-1]
		cmdArgs := buildArgs(number, total, sockets, prog.args)

		cmd := exec.Command(prog.name, cmdArgs...)
		commands[number] = cmd

		prefix := fmt.Sprintf("%s_%d", filepath.Base(prog.name), number)
		attachOutput(cmd, prefix)

		if err := cmd.Start(); err != nil {
			fmt.Fprintln(os.Stderr, "chain: could not start", prog.name, err)
			triggerStop()
			break
		}

		running.Add(1)
		go func(cmd *exec.Cmd) {
			defer running.Done()
			cmd.Wait()
			triggerStop()
		}(cmd)

		if number > 1 {
			if !waitForFile(sockets[number], stop) {
				break
			}
		}
	}

	<-stop

	for number := total; number >= 1; number-- {
		cmd := commands[number]
		if cmd != nil && cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	running.Wait()
}

func waitForFile(path string, stop <-chan struct{}) bool {
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	for {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		select {
		case <-stop:
			return false
		case <-ticker.C:
		}
	}
}

func attachOutput(cmd *exec.Cmd, prefix string) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	go printLines(stdout, prefix, os.Stdout)
	go printLines(stderr, prefix, os.Stderr)
}

func printLines(reader io.Reader, prefix string, out io.Writer) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		fmt.Fprintln(out, prefix, scanner.Text())
	}
}

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"
)

// util
func dieIfErr(e error, reason string) {
	if e != nil {
		log.Fatalln(reason + ": " + e.Error())
	}
}

// util
func splitArray(array []string, separator string) [][]string {
	shouldMakeNewGroup := true
	groups := [][]string{}
	for _, item := range array {
		if item == separator {
			shouldMakeNewGroup = true
		} else {
			if shouldMakeNewGroup {
				groups = append(groups, []string{})
				shouldMakeNewGroup = false
			}
			groups[len(groups) - 1] = append(groups[len(groups) - 1], item)
		}
	}
	return groups
}

func getArgs() (string, [][]string) {
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatalln("Usage: chain ./socket_storage_dir SEPARATOR subprogram arg arg SEPARATOR subprogram arg arg SEPARATOR ...")
	}
	socketStorageDir := args[0]
	separator := args[1]
	subprograms := splitArray(args[2:], separator)
	return socketStorageDir, subprograms
}

func prepare(socketStorageDir string) {
	os.MkdirAll(socketStorageDir, 0755)
	entries, err := os.ReadDir(socketStorageDir)
	dieIfErr(err, "couldn't read socket storage dir")
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".socket") {
			socketPath := path.Join(socketStorageDir, entry.Name())
			dieIfErr(os.Remove(socketPath), fmt.Sprint("couldn't remove socket %v", socketPath))
		}
	}
}

func run(socketStorageDir string, subprograms [][]string) {
	var exitedSubprogramNumbers chan int
	for n := len(subprograms); n > 0; n-- {
		subprogram := subprograms[n - 1]
		cmd := exec.Command(subprogram[0], subprogram[1:]...)
		cmd.Env = os.Environ()
		input := ""
		if n != len(subprograms) {
			cmd.Env = append(cmd.Env, "OUTPUT=" + path.Join(socketStorageDir, fmt.Sprint(n + 1, ".socket")))
		}
		if n != 1 {
			input = path.Join(socketStorageDir, fmt.Sprint(n, ".socket"))
			cmd.Env = append(cmd.Env, "INPUT=" + input)
		}
		stdoutPipe, err := cmd.StdoutPipe()
		dieIfErr(err, fmt.Sprintf("couldn't get stdout pipe for subprogram %v", n))
		stderrPipe, err := cmd.StderrPipe()
		dieIfErr(err, fmt.Sprintf("couldn't get stderr pipe for subprogram %v", n))
		dieIfErr(cmd.Start(), fmt.Sprintf("couldn't start subprogram %v", n))
		copyWithPrefix := func(destination io.Writer, source io.Reader) {
			scanner := bufio.NewScanner(source)
			for scanner.Scan() {
				fmt.Fprint(destination, n, ": ", scanner.Text(), "\n")
			}
		}
		go copyWithPrefix(os.Stdout, stdoutPipe)
		go copyWithPrefix(os.Stderr, stderrPipe)
		if input != "" {
			for {
				if _, err := os.Stat(input); err != nil {
					break
				}
				time.Sleep(time.Millisecond * 50)
			}
		}
		go func () {
			cmd.Wait()
			exitedSubprogramNumbers <- n
		}()
	}
	log.Fatalln("subprogram", <- exitedSubprogramNumbers, "exited")
}

func main() {
	socketStorageDir, subprograms := getArgs()
	prepare(socketStorageDir)
	run(socketStorageDir, subprograms)
}

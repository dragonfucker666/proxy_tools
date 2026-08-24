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
	"sync"
	"time"
)

// util
func dieIfErr(e error) {
	if e != nil {
		log.Fatalln(e)
	}
}

// util
func splitArray(array []string, separator string) [][]string {
	shouldMakeNewGroup := true
	groups := make([][]string, 0)
	for _, item := range array {
		if item == separator {
			shouldMakeNewGroup = true
		} else {
			if shouldMakeNewGroup {
				groups = append(groups, make([]string, 0))
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
	dieIfErr(err)
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".socket") {
			dieIfErr(os.Remove(path.Join(socketStorageDir, entry.Name())))
		}
	}
}

func run(socketStorageDir string, subprograms [][]string) {
	var wg sync.WaitGroup
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
		dieIfErr(err)
		stderrPipe, err := cmd.StderrPipe()
		dieIfErr(err)
		dieIfErr(cmd.Start())
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
		wg.Add(1)
		go func () {
			dieIfErr(cmd.Wait())
			wg.Done()
		}()
	}
	wg.Wait()
}

func main() {
	socketStorageDir, subprograms := getArgs()
	prepare(socketStorageDir)
	run(socketStorageDir, subprograms)
}

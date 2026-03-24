package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var validKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	var kvs map[string]string

	if err := json.Unmarshal(input, &kvs); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for k := range kvs {
		if !validKey.MatchString(k) {
			fmt.Fprintf(os.Stderr, "invalid key: %q\n", k)
			os.Exit(1)
		}
	}

	// Check for exec mode
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--" {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "expected command after --\n")
			os.Exit(1)
		}

		cmd := args[1]
		cmdArgs := args[2:]

		for k, v := range kvs {
			os.Setenv(k, v)
		}

		binary, err := exec.LookPath(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not find command: %v\n", err)
			os.Exit(1)
		}

		runExec(binary, cmd, cmdArgs)
		return
	}

	// Default: print export statements
	for k, v := range kvs {
		fmt.Printf("export %s='%s'\n", k, strings.ReplaceAll(v, "'", "'\\''"))
	}
}

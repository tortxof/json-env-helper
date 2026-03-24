//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func runExec(binary, cmd string, cmdArgs []string) {
	c := exec.Command(binary, cmdArgs...)
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

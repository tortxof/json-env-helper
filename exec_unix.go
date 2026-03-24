//go:build !windows

package main

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func runExec(binary, cmd string, cmdArgs []string) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: could not open /dev/tty: %v\n", err)
	} else {
		if err := unix.Dup2(int(tty.Fd()), int(os.Stdin.Fd())); err != nil {
			fmt.Fprintf(os.Stderr, "warning: could not reopen stdin: %v\n", err)
		}
		tty.Close()
	}

	if err := syscall.Exec(binary, append([]string{cmd}, cmdArgs...), os.Environ()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

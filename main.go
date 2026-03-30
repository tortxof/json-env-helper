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

type credentialProcessOutput struct {
	Version         int    `json:"Version"`
	AccessKeyId     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	SessionToken    string `json:"SessionToken"`
	Expiration      string `json:"Expiration"`
}

func parseArgs(osArgs []string) (credentialProcess bool, execMode bool, cmd string, cmdArgs []string, err error) {
	i := 0
	for i < len(osArgs) {
		arg := osArgs[i]
		if arg == "--credential-process" {
			credentialProcess = true
			i++
			continue
		}
		if arg == "--" {
			execMode = true
			rest := osArgs[i+1:]
			if len(rest) == 0 {
				err = fmt.Errorf("expected command after --")
				return
			}
			cmd = rest[0]
			cmdArgs = rest[1:]
			return
		}
		// Unknown flag or positional arg
		err = fmt.Errorf("unexpected argument: %q", arg)
		return
	}
	return
}

func parseInput(input []byte, credentialProcess bool) (map[string]string, error) {
	if credentialProcess {
		var cpo credentialProcessOutput
		if err := json.Unmarshal(input, &cpo); err != nil {
			return nil, fmt.Errorf("error parsing credential_process JSON: %v", err)
		}
		if cpo.Version != 1 {
			return nil, fmt.Errorf("unsupported credential_process Version: %d (expected 1)", cpo.Version)
		}
		if cpo.AccessKeyId == "" {
			return nil, fmt.Errorf("missing required field: AccessKeyId")
		}
		if cpo.SecretAccessKey == "" {
			return nil, fmt.Errorf("missing required field: SecretAccessKey")
		}
		kvs := map[string]string{
			"AWS_ACCESS_KEY_ID":     cpo.AccessKeyId,
			"AWS_SECRET_ACCESS_KEY": cpo.SecretAccessKey,
		}
		if cpo.SessionToken != "" {
			kvs["AWS_SESSION_TOKEN"] = cpo.SessionToken
		}
		if cpo.Expiration != "" {
			kvs["AWS_CREDENTIAL_EXPIRATION"] = cpo.Expiration
		}
		return kvs, nil
	}

	var kvs map[string]string
	if err := json.Unmarshal(input, &kvs); err != nil {
		return nil, fmt.Errorf("error: %v", err)
	}
	for k := range kvs {
		if !validKey.MatchString(k) {
			return nil, fmt.Errorf("invalid key: %q", k)
		}
	}
	return kvs, nil
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	credentialProcess, execMode, cmd, cmdArgs, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	kvs, err := parseInput(input, credentialProcess)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if execMode {
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

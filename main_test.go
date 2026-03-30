package main

import (
	"testing"
)

// --- parseArgs tests ---

func TestParseArgs_NoArgs(t *testing.T) {
	cp, exec, cmd, cmdArgs, err := parseArgs([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp || exec || cmd != "" || len(cmdArgs) != 0 {
		t.Fatalf("expected zero values, got cp=%v exec=%v cmd=%q cmdArgs=%v", cp, exec, cmd, cmdArgs)
	}
}

func TestParseArgs_CredentialProcessFlag(t *testing.T) {
	cp, exec, cmd, _, err := parseArgs([]string{"--credential-process"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cp {
		t.Error("expected credentialProcess=true")
	}
	if exec || cmd != "" {
		t.Error("expected no exec mode")
	}
}

func TestParseArgs_ExecMode(t *testing.T) {
	cp, execMode, cmd, cmdArgs, err := parseArgs([]string{"--", "myapp", "arg1", "arg2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp {
		t.Error("expected credentialProcess=false")
	}
	if !execMode {
		t.Error("expected execMode=true")
	}
	if cmd != "myapp" {
		t.Errorf("expected cmd=myapp, got %q", cmd)
	}
	if len(cmdArgs) != 2 || cmdArgs[0] != "arg1" || cmdArgs[1] != "arg2" {
		t.Errorf("unexpected cmdArgs: %v", cmdArgs)
	}
}

func TestParseArgs_CredentialProcessBeforeExec(t *testing.T) {
	cp, execMode, cmd, cmdArgs, err := parseArgs([]string{"--credential-process", "--", "myapp", "arg1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cp {
		t.Error("expected credentialProcess=true")
	}
	if !execMode {
		t.Error("expected execMode=true")
	}
	if cmd != "myapp" {
		t.Errorf("expected cmd=myapp, got %q", cmd)
	}
	if len(cmdArgs) != 1 || cmdArgs[0] != "arg1" {
		t.Errorf("unexpected cmdArgs: %v", cmdArgs)
	}
}

func TestParseArgs_ExecModeNoCommand(t *testing.T) {
	_, _, _, _, err := parseArgs([]string{"--"})
	if err == nil {
		t.Fatal("expected error for missing command after --")
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	_, _, _, _, err := parseArgs([]string{"--unknown"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
}

// --- parseInput tests: flat map mode ---

func TestParseInput_FlatMap(t *testing.T) {
	input := []byte(`{"FOO": "bar", "BAZ": "qux"}`)
	kvs, err := parseInput(input, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kvs["FOO"] != "bar" || kvs["BAZ"] != "qux" {
		t.Errorf("unexpected kvs: %v", kvs)
	}
}

func TestParseInput_FlatMap_InvalidKey(t *testing.T) {
	input := []byte(`{"invalid-key": "value"}`)
	_, err := parseInput(input, false)
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
}

func TestParseInput_FlatMap_InvalidJSON(t *testing.T) {
	input := []byte(`not json`)
	_, err := parseInput(input, false)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseInput_FlatMap_NonStringValue(t *testing.T) {
	// map[string]string will reject non-string values
	input := []byte(`{"FOO": 123}`)
	_, err := parseInput(input, false)
	if err == nil {
		t.Fatal("expected error for non-string value")
	}
}

// --- parseInput tests: credential_process mode ---

func TestParseInput_CredentialProcess_Valid(t *testing.T) {
	input := []byte(`{
		"Version": 1,
		"AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
		"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"SessionToken": "AQoDYXdzEJr...",
		"Expiration": "2024-01-01T00:00:00Z"
	}`)
	kvs, err := parseInput(input, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kvs["AWS_ACCESS_KEY_ID"] != "AKIAIOSFODNN7EXAMPLE" {
		t.Errorf("unexpected AWS_ACCESS_KEY_ID: %q", kvs["AWS_ACCESS_KEY_ID"])
	}
	if kvs["AWS_SECRET_ACCESS_KEY"] != "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" {
		t.Errorf("unexpected AWS_SECRET_ACCESS_KEY: %q", kvs["AWS_SECRET_ACCESS_KEY"])
	}
	if kvs["AWS_SESSION_TOKEN"] != "AQoDYXdzEJr..." {
		t.Errorf("unexpected AWS_SESSION_TOKEN: %q", kvs["AWS_SESSION_TOKEN"])
	}
	if kvs["AWS_CREDENTIAL_EXPIRATION"] != "2024-01-01T00:00:00Z" {
		t.Errorf("unexpected AWS_CREDENTIAL_EXPIRATION: %q", kvs["AWS_CREDENTIAL_EXPIRATION"])
	}
}

func TestParseInput_CredentialProcess_OptionalFieldsAbsent(t *testing.T) {
	input := []byte(`{
		"Version": 1,
		"AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
		"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	}`)
	kvs, err := parseInput(input, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := kvs["AWS_SESSION_TOKEN"]; ok {
		t.Error("AWS_SESSION_TOKEN should not be present when SessionToken is absent")
	}
	if _, ok := kvs["AWS_CREDENTIAL_EXPIRATION"]; ok {
		t.Error("AWS_CREDENTIAL_EXPIRATION should not be present when Expiration is absent")
	}
}

func TestParseInput_CredentialProcess_WrongVersion(t *testing.T) {
	input := []byte(`{
		"Version": 2,
		"AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
		"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	}`)
	_, err := parseInput(input, true)
	if err == nil {
		t.Fatal("expected error for unsupported Version")
	}
}

func TestParseInput_CredentialProcess_MissingAccessKeyId(t *testing.T) {
	input := []byte(`{
		"Version": 1,
		"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	}`)
	_, err := parseInput(input, true)
	if err == nil {
		t.Fatal("expected error for missing AccessKeyId")
	}
}

func TestParseInput_CredentialProcess_MissingSecretAccessKey(t *testing.T) {
	input := []byte(`{
		"Version": 1,
		"AccessKeyId": "AKIAIOSFODNN7EXAMPLE"
	}`)
	_, err := parseInput(input, true)
	if err == nil {
		t.Fatal("expected error for missing SecretAccessKey")
	}
}

func TestParseInput_CredentialProcess_InvalidJSON(t *testing.T) {
	input := []byte(`not json`)
	_, err := parseInput(input, true)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseInput_CredentialProcess_UnknownFieldsIgnored(t *testing.T) {
	// Unknown fields should be silently ignored
	input := []byte(`{
		"Version": 1,
		"AccessKeyId": "AKIAIOSFODNN7EXAMPLE",
		"SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		"SomeFutureField": "value"
	}`)
	_, err := parseInput(input, true)
	if err != nil {
		t.Fatalf("unexpected error for unknown field: %v", err)
	}
}

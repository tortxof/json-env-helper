# json-env-helper

A small CLI helper that reads JSON key/value pairs from stdin and outputs shell
export statements or executes a command with the environment variables set.

This allows secrets to be bulk-loaded into your shell session without ever
writing them to disk.

## Installation

### Mise

```bash
mise use github:tortxof/json-env-helper
```

### Go

```bash
go install github.com/tortxof/json-env-helper@latest
```

## Usage

### Eval mode

Outputs shell export statements to be evaluated by the current shell:

With AWS Secrets Manager:
```
eval "$(aws secretsmanager get-secret-value --secret-id my/secret --query SecretString --output text | json-env-helper)"
```

### Exec mode

Executes a command with the environment variables set:

With AWS Secrets Manager:
```
aws secretsmanager get-secret-value --secret-id my/secret --query SecretString --output text | json-env-helper -- myapp --flag arg
```

With direct JSON input:
```
echo '{"DB_HOST": "localhost", "DB_PORT": "5432"}' | json-env-helper -- myapp --flag arg
```

## Input Format

The input must be a JSON object with string keys and string values:

```json
{
  "DB_HOST": "localhost",
  "DB_PORT": "5432",
  "API_KEY": "supersecret"
}
```

Keys must be valid environment variable names: they must start with a letter or
underscore, and contain only letters, digits, and underscores (`[A-Za-z_][A-Za-z0-9_]*`).
Any key that does not match this pattern will cause the program to exit with an error.

## Security

- Secrets are never written to disk.
- In **Eval mode**, secrets exist only in your current shell session's memory.
- In **Exec mode**, secrets exist only in the spawned process's environment.
- Keys are validated to contain only letters, digits, and underscores, preventing shell injection via key names.
- Values are single-quoted in the output, preventing shell expansion of `$`, backticks, and `$(...)` in eval mode.

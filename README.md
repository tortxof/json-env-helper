# json-env-helper

A small CLI helper that reads JSON from stdin and outputs shell export
statements or executes a command with the environment variables set. Supports
both arbitrary key/value JSON and AWS `credential_process` output format.

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

The input must be a flat JSON object with string keys and string values, or JSON
in AWS credential process format.

### Flat JSON format

```json
{
  "DB_HOST": "localhost",
  "DB_PORT": "5432",
  "API_KEY": "supersecret"
}
```

Keys must be valid environment variable names: they must start with a letter or
underscore, and contain only letters, digits, and underscores
(`[A-Za-z_][A-Za-z0-9_]*`). Any key that does not match this pattern will cause
the program to exit with an error.

### Credential process input

The `--credential-process` flag enables parsing of AWS `credential_process` JSON
output. This is useful when using an AWS credential helper that outputs
credentials in the [credential process
format](https://docs.aws.amazon.com/sdkref/latest/guide/feature-process-credentials.html).

The flag can be combined with eval mode or exec mode. Useful for software that
strictly relies on credentials being present in environment variables, and
doesn't support `aws login`.

```bash
aws configure export-credentials --format process | json-env-helper --credential-process -- restic backup ~
```

The input is mapped to the following environment variables:

| JSON field        | Environment variable        | Required |
|-------------------|-----------------------------|----------|
| `AccessKeyId`     | `AWS_ACCESS_KEY_ID`         | yes      |
| `SecretAccessKey` | `AWS_SECRET_ACCESS_KEY`     | yes      |
| `SessionToken`    | `AWS_SESSION_TOKEN`         | no       |
| `Expiration`      | `AWS_CREDENTIAL_EXPIRATION` | no       |

The `Version` field must be present and set to `1`.

## Security

- Secrets are never written to disk.
- In **Eval mode**, secrets exist only in your current shell session's memory.
- In **Exec mode**, secrets exist only in the spawned process's environment.
- In standard mode, keys are validated to contain only letters, digits, and
  underscores, preventing shell injection via key names. In
  `--credential-process` mode, output keys are hardcoded AWS variable names.
- Values are single-quoted in the output, preventing shell expansion of `$`,
  backticks, and `$(...)` in eval mode.

# kinmokusei-env

`kinmokusei-env` loads configuration from `.env` files and process environment
variables through a Kinmokusei-first API. It keeps parsing, type conversion,
and process mutation explicit so applications remain easy to test.

## Requirements

- Kinmokusei v0.4.3 or later
- Go 1.23 or later

## Features

- Parse `.env` text or load a file
- Unquoted, single-quoted, and double-quoted values
- Double-quoted `\n`, `\r`, `\t`, `\\`, and `\"` escapes
- Full-line and whitespace-delimited inline comments
- Optional `export` prefixes, UTF-8 BOM, and CRLF input
- Required text and typed integer/boolean accessors
- Defaults that only apply when a variable is absent
- Import selected process variables
- Apply values while preserving or overwriting existing process variables
- Errors returned through `Result<T>` rather than process termination

Variable expansion is intentionally not performed. Values such as `$HOME` and
`${TOKEN}` remain literal, avoiding implicit dependencies on the calling
process environment.

## Installation

```sh
keika deps add github.com/puffball1567/kinmokusei-env@v0.1.0
```

Kinmokusei registers the package's short import automatically:

```ts
import { load } from "kinmokusei-env";
```

## Example

Given this `.env` file:

```dotenv
HOST=127.0.0.1
PORT=7000
DEBUG=true
```

Load and validate it in Kinmokusei:

```ts
import { load } from "kinmokusei-env";
import go fmt from "fmt";

function main(): void {
  const [environment, loadError] = load(".env");
  if (loadError !== nil) {
    fmt.Println("error:", loadError.Error());
    return;
  }

  const host = environment.textOr("HOST", "localhost");
  const [port, portError] = environment.integerOr("PORT", 8080);
  if (portError !== nil) {
    fmt.Println("error:", portError.Error());
    return;
  }
  const [debug, debugError] = environment.booleanOr("DEBUG", false);
  if (debugError !== nil) {
    fmt.Println("error:", debugError.Error());
    return;
  }

  fmt.Printf("%s:%d debug=%t\n", host, port, debug);
}
```

A present but empty value is different from an absent value. Defaults are used
only for absent variables. Invalid present integer or boolean values return an
error instead of silently using a default.

## Process environment

`fromProcess(names)` copies only the requested variables into an independent
`Environment`. Missing variables are omitted. `apply(false)` fills variables
that are currently absent, while `apply(true)` replaces existing values.

```ts
const environment = parse("LOG_LEVEL=debug\n")?;
environment.apply(false)?;
```

## Parsing behavior

- Names use the portable pattern `[A-Za-z_][A-Za-z0-9_]*`.
- Leading and trailing whitespace around names and unquoted values is removed.
- `#` starts an inline comment in an unquoted value only when preceded by a
  space or tab. It is literal inside quotes and URLs such as `/#fragment`.
- Single-quoted values are literal. Double-quoted values process the supported
  escapes listed above.
- Text after a closing quote must be whitespace or a comment.
- If a name occurs more than once, the last assignment wins.
- Empty assignments such as `NAME=` are valid.

## Development

Run source checks, the parser/accessor matrix, 100% generated statement
coverage, fuzz tests, the example, and an independent short-import consumer:

```sh
KEIKA=keika sh scripts/test.sh
```

Generated Go is test output and is not maintained as source.

## License

Apache-2.0

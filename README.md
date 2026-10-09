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
- Floating-point, non-empty, choice, and inclusive integer-range validation
- Defaults that only apply when a variable is absent
- Sorted name enumeration, removal, cloning, and layered merging
- Strict duplicate-key parsing when configuration mistakes must fail
- Deterministic `.env` encoding with round-trip-safe quoting
- Explicit `${NAME}` expansion with missing-reference and cycle detection
- Import selected variables or snapshot the complete process environment
- Apply values while preserving or overwriting existing process variables
- Optional file loading for local overrides
- Errors returned through `Result<T>` rather than process termination

Variable expansion is explicit. Normal parsing leaves `$HOME` and `${TOKEN}`
literal; call `expanded()` only when references should be resolved from the
same `Environment`. Process variables are never consulted implicitly.

## Installation

```sh
keika deps add github.com/puffball1567/kinmokusei-env@v0.2.0
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
`fromProcessAll()` snapshots every process variable whose name is portable.

See [`examples/layered/main.km`](examples/layered/main.km) for a complete
defaults → `.env` file → process environment composition example.

```ts
const environment = parse("LOG_LEVEL=debug\n")?;
environment.apply(false)?;
```

## Layered configuration

Use `merge(other, overwrite)` to build the common defaults → file → process
configuration stack without mutating the source layers. `clone()` creates the
initial independent layer.

```ts
const combined = defaults.clone();
combined.merge(fileValues, true);
combined.merge(processValues, true);
combined.require(["DATABASE_URL", "SECRET_KEY"])?;
const port = combined.integerRange("PORT", 1, 65535)?;
const mode = combined.choice("MODE", ["development", "production"])?;
```

`names()` is sorted for reproducible diagnostics. `encode()` uses that order
and quotes every value so parsing the output reconstructs the same data.

## Strict parsing and expansion

`parse()` accepts repeated names and keeps the last assignment, which is handy
for layered text assembled by another tool. `parseStrict()` rejects repeats
with their line number. `loadStrict()` provides the same policy for files;
`loadIfExists()` returns an empty environment only when the file does not
exist and still reports every other I/O or parse error.

`expanded()` returns a new environment and recursively resolves `${NAME}`
references from that environment. It reports undefined references, malformed
names, unterminated references, and direct or indirect cycles. The original
environment is unchanged, and `$NAME` without braces remains literal.

## Parsing behavior

- Names use the portable pattern `[A-Za-z_][A-Za-z0-9_]*`.
- Leading and trailing whitespace around names and unquoted values is removed.
- `#` starts an inline comment in an unquoted value only when preceded by a
  space or tab. It is literal inside quotes and URLs such as `/#fragment`.
- Single-quoted values are literal. Double-quoted values process the supported
  escapes listed above.
- Text after a closing quote must be whitespace or a comment.
- If a name occurs more than once, `parse()` keeps the last assignment and
  `parseStrict()` reports an error.
- Empty assignments such as `NAME=` are valid.

## Development

Run source checks, the complete parser/accessor/layering/expansion matrix, 100%
generated statement coverage, four fuzz targets, the example, and an
independent short-import consumer:

```sh
KEIKA=keika sh scripts/test.sh
```

Generated Go is test output and is not maintained as source.

## License

Apache-2.0

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development, pull request,
release, and hotfix rules.

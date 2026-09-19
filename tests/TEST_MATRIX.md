# Test matrix

| Area | Normal cases | Edge cases | Failure cases |
| --- | --- | --- | --- |
| Lines | assignments, `export`, comments | blank lines, BOM, CRLF, Unicode values, duplicate keys | missing `=`, invalid names |
| Values | unquoted, single quoted, double quoted | empty values, embedded `#`, whitespace, supported escapes | unterminated quotes, unsupported escapes, trailing text |
| Typed access | text, integer, boolean | negative integer, empty text, defaults, explicit false | missing required value, invalid integer/boolean |
| Advanced validation | number, non-empty text, choice, integer range, required set | inclusive range boundaries, `NaN`, empty required set | empty choices, reversed range, disallowed/out-of-range values |
| Collections | names, size, clone, remove, merge | deterministic order, preserve/overwrite modes, empty environment | missing removal |
| Files | normal/strict/optional loading | absent optional file, empty/trailing newline | missing required file, directory read, strict duplicate |
| Process environment | import present variables, apply values | missing variables, preserve/overwrite modes | invalid names, OS rejection |
| Encoding | deterministic `.env` output | empty, control characters, quotes, slashes, Unicode | arbitrary-value round-trip fuzzing |
| Expansion | direct, forward, repeated, nested references | literal `$NAME`, source immutability, cached dependency | undefined/invalid names, unterminated reference, direct/indirect cycles |
| Robustness | fixed regression corpus | arbitrary text/name/value/expansion fuzzing | parsers, accessors, encoders, and expansion must never panic |

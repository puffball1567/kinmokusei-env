# Test matrix

| Area | Normal cases | Edge cases | Failure cases |
| --- | --- | --- | --- |
| Lines | assignments, `export`, comments | blank lines, BOM, CRLF, Unicode values, duplicate keys | missing `=`, invalid names |
| Values | unquoted, single quoted, double quoted | empty values, embedded `#`, whitespace, supported escapes | unterminated quotes, unsupported escapes, trailing text |
| Typed access | text, integer, boolean | negative integer, empty text, defaults, explicit false | missing required value, invalid integer/boolean |
| Files | readable `.env` file | empty/trailing newline handled by parser | missing file |
| Process environment | import present variables, apply values | missing variables, preserve/overwrite modes | invalid names, OS rejection |
| Robustness | fixed regression corpus | arbitrary text/name/value fuzzing | parsers and accessors must never panic |

package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func mustParse(t *testing.T, source string) *Environment {
	t.Helper()
	got, err := parse(source)
	if err != nil {
		t.Fatalf("parse() error = %v", err)
	}
	return got
}

func TestParseSyntaxMatrix(t *testing.T) {
	tests := []struct {
		name   string
		source string
		key    string
		want   string
	}{
		{"empty value", "EMPTY=", "EMPTY", ""},
		{"outer whitespace", "  NAME = value  ", "NAME", "value"},
		{"export space", "export NAME=value", "NAME", "value"},
		{"export tab", "export\tNAME=value", "NAME", "value"},
		{"hash without whitespace", "URL=https://example.test/#part", "URL", "https://example.test/#part"},
		{"inline comment with space", "NAME=value # comment", "NAME", "value"},
		{"inline comment with tab", "NAME=value\t# comment", "NAME", "value"},
		{"single quoted literal", "NAME='a\\nb # c'", "NAME", `a\nb # c`},
		{"single quoted comment", "NAME='value' # comment", "NAME", "value"},
		{"double quoted escapes", `NAME="a\n\r\t\\\"b"`, "NAME", "a\n\r\t\\\"b"},
		{"double quoted comment", `NAME="value" # comment`, "NAME", "value"},
		{"unicode", "NAME=金木犀", "NAME", "金木犀"},
		{"bom", "\ufeffNAME=value", "NAME", "value"},
		{"crlf", "NAME=value\r\nOTHER=second\r\n", "OTHER", "second"},
		{"underscore name", "_NAME_2=value", "_NAME_2", "value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			environment := mustParse(t, test.source)
			got, err := environment.Text(test.key)
			if err != nil || got != test.want {
				t.Fatalf("Text(%q) = (%q, %v), want %q", test.key, got, err, test.want)
			}
		})
	}
}

func TestBlankLinesCommentsAndDuplicates(t *testing.T) {
	environment := mustParse(t, "\n # ignored\nVALUE=first\nVALUE=second\n\t\n")
	if got := environment.TextOr("VALUE", "missing"); got != "second" {
		t.Fatalf("TextOr(VALUE) = %q, want second", got)
	}
	if got := environment.TextOr("MISSING", "fallback"); got != "fallback" {
		t.Fatalf("TextOr(MISSING) = %q, want fallback", got)
	}
	if environment.Has("MISSING") || !environment.Has("VALUE") {
		t.Fatalf("Has() returned unexpected results")
	}
}

func TestParseErrorMatrix(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"missing equals", "NAME", "line 1: expected NAME=VALUE"},
		{"empty name", "=value", `invalid environment variable name ""`},
		{"starts with digit", "1NAME=value", `invalid environment variable name "1NAME"`},
		{"hyphenated name", "BAD-NAME=value", `invalid environment variable name "BAD-NAME"`},
		{"unicode name", "名前=value", `invalid environment variable name "名前"`},
		{"empty export", "export =value", `invalid environment variable name ""`},
		{"unterminated single quote", "NAME='value", "unterminated single-quoted value"},
		{"unterminated double quote", `NAME="value`, "unterminated double-quoted value"},
		{"dangling escape", `NAME="value\`, "unterminated double-quoted value"},
		{"unsupported escape", `NAME="a\q"`, `unsupported escape sequence \q`},
		{"text after single quote", "NAME='value' suffix", "unexpected text after quoted value"},
		{"text after double quote", `NAME="value" suffix`, "unexpected text after quoted value"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := parse(test.source)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("parse(%q) error = %v, want substring %q", test.source, err, test.want)
			}
		})
	}
}

func TestSetAndTypedAccessors(t *testing.T) {
	environment := NewEnvironment()
	values := map[string]string{
		"TEXT":        "hello",
		"EMPTY":       "",
		"INTEGER":     "-42",
		"BAD_INTEGER": "4.2",
		"TRUE":        "true",
		"FALSE":       "0",
		"BAD_BOOLEAN": "sometimes",
	}
	for name, value := range values {
		if err := environment.Set(name, value); err != nil {
			t.Fatalf("Set(%q) error = %v", name, err)
		}
	}

	if got, err := environment.Text("EMPTY"); err != nil || got != "" {
		t.Fatalf("Text(EMPTY) = (%q, %v)", got, err)
	}
	if _, err := environment.Text("MISSING"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("Text(MISSING) error = %v", err)
	}
	if got, err := environment.Integer("INTEGER"); err != nil || got != -42 {
		t.Fatalf("Integer(INTEGER) = (%d, %v)", got, err)
	}
	if _, err := environment.Integer("BAD_INTEGER"); err == nil || !strings.Contains(err.Error(), "invalid integer") {
		t.Fatalf("Integer(BAD_INTEGER) error = %v", err)
	}
	if _, err := environment.Integer("MISSING"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("Integer(MISSING) error = %v", err)
	}
	if got, err := environment.IntegerOr("MISSING", 17); err != nil || got != 17 {
		t.Fatalf("IntegerOr(MISSING) = (%d, %v)", got, err)
	}
	if got, err := environment.IntegerOr("INTEGER", 17); err != nil || got != -42 {
		t.Fatalf("IntegerOr(INTEGER) = (%d, %v)", got, err)
	}
	if _, err := environment.IntegerOr("BAD_INTEGER", 17); err == nil {
		t.Fatalf("IntegerOr(BAD_INTEGER) error = nil")
	}

	if got, err := environment.Boolean("TRUE"); err != nil || !got {
		t.Fatalf("Boolean(TRUE) = (%v, %v)", got, err)
	}
	if got, err := environment.Boolean("FALSE"); err != nil || got {
		t.Fatalf("Boolean(FALSE) = (%v, %v)", got, err)
	}
	if _, err := environment.Boolean("BAD_BOOLEAN"); err == nil || !strings.Contains(err.Error(), "invalid boolean") {
		t.Fatalf("Boolean(BAD_BOOLEAN) error = %v", err)
	}
	if _, err := environment.Boolean("MISSING"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("Boolean(MISSING) error = %v", err)
	}
	if got, err := environment.BooleanOr("MISSING", true); err != nil || !got {
		t.Fatalf("BooleanOr(MISSING) = (%v, %v)", got, err)
	}
	if got, err := environment.BooleanOr("FALSE", true); err != nil || got {
		t.Fatalf("BooleanOr(FALSE) = (%v, %v)", got, err)
	}
	if _, err := environment.BooleanOr("BAD_BOOLEAN", true); err == nil {
		t.Fatalf("BooleanOr(BAD_BOOLEAN) error = nil")
	}
}

func TestSetRejectsInvalidNames(t *testing.T) {
	for _, name := range []string{"", "1NAME", "BAD-NAME", "BAD NAME", "名前", "A=B"} {
		if err := NewEnvironment().Set(name, "value"); err == nil {
			t.Errorf("Set(%q) error = nil", name)
		}
	}
}

func TestLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("NAME=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	environment, err := load(path)
	if err != nil || environment.TextOr("NAME", "") != "from-file" {
		t.Fatalf("load() = (%#v, %v)", environment, err)
	}
	if _, err := load(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatalf("load(missing) error = nil")
	}
}

func TestFromProcess(t *testing.T) {
	const present = "KINMOKUSEI_ENV_TEST_PRESENT"
	const missing = "KINMOKUSEI_ENV_TEST_MISSING"
	t.Setenv(present, "value")
	os.Unsetenv(missing)

	environment, err := fromProcess([]string{present, missing})
	if err != nil {
		t.Fatalf("fromProcess() error = %v", err)
	}
	if got := environment.TextOr(present, ""); got != "value" || environment.Has(missing) {
		t.Fatalf("fromProcess() values = (%q, missing=%v)", got, environment.Has(missing))
	}
	if _, err := fromProcess([]string{"BAD-NAME"}); err == nil {
		t.Fatalf("fromProcess(invalid) error = nil")
	}
}

func TestApplyOverwriteMatrix(t *testing.T) {
	const key = "KINMOKUSEI_ENV_TEST_APPLY"

	t.Run("sets missing", func(t *testing.T) {
		os.Unsetenv(key)
		t.Cleanup(func() { os.Unsetenv(key) })
		environment := mustParse(t, key+"=new")
		if err := environment.Apply(false); err != nil || os.Getenv(key) != "new" {
			t.Fatalf("Apply(false) = %v, value %q", err, os.Getenv(key))
		}
	})

	t.Run("preserves existing", func(t *testing.T) {
		t.Setenv(key, "old")
		environment := mustParse(t, key+"=new")
		if err := environment.Apply(false); err != nil || os.Getenv(key) != "old" {
			t.Fatalf("Apply(false) = %v, value %q", err, os.Getenv(key))
		}
	})

	t.Run("overwrites existing", func(t *testing.T) {
		t.Setenv(key, "old")
		environment := mustParse(t, key+"=new")
		if err := environment.Apply(true); err != nil || os.Getenv(key) != "new" {
			t.Fatalf("Apply(true) = %v, value %q", err, os.Getenv(key))
		}
	})

	t.Run("propagates operating system error", func(t *testing.T) {
		environment := NewEnvironment()
		environment.values["BAD=NAME"] = "value"
		if err := environment.Apply(true); err == nil {
			t.Fatalf("Apply(true) error = nil")
		}
	})
}

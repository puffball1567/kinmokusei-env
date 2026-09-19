package env

import (
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCollectionOperations(t *testing.T) {
	empty := NewEnvironment()
	if empty.Size() != 0 || len(empty.Names()) != 0 || empty.Encode() != "" {
		t.Fatalf("new environment is not empty")
	}

	environment := mustParse(t, "Z=last\nA=first\nM=middle\n")
	if got, want := environment.Size(), 3; got != want {
		t.Fatalf("Size() = %d, want %d", got, want)
	}
	if got, want := environment.Names(), []string{"A", "M", "Z"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Names() = %#v, want %#v", got, want)
	}
	if environment.Remove("MISSING") {
		t.Fatalf("Remove(MISSING) = true")
	}
	if !environment.Remove("M") || environment.Has("M") || environment.Size() != 2 {
		t.Fatalf("Remove(M) did not remove the value")
	}

	cloned := environment.Clone()
	if err := cloned.Set("A", "changed"); err != nil {
		t.Fatal(err)
	}
	if cloned.Remove("Z") != true || environment.TextOr("A", "") != "first" || !environment.Has("Z") {
		t.Fatalf("Clone() shares mutable state with the source")
	}
}

func TestMergeOverwriteMatrix(t *testing.T) {
	base := mustParse(t, "SHARED=base\nBASE_ONLY=yes\n")
	overlay := mustParse(t, "SHARED=overlay\nOVERLAY_ONLY=yes\n")

	preserved := base.Clone()
	preserved.Merge(overlay, false)
	if preserved.TextOr("SHARED", "") != "base" || preserved.TextOr("OVERLAY_ONLY", "") != "yes" {
		t.Fatalf("Merge(false) produced unexpected values")
	}

	overwritten := base.Clone()
	overwritten.Merge(overlay, true)
	if overwritten.TextOr("SHARED", "") != "overlay" || overwritten.TextOr("BASE_ONLY", "") != "yes" {
		t.Fatalf("Merge(true) produced unexpected values")
	}
}

func TestAdvancedTypedAccessors(t *testing.T) {
	environment := mustParse(t, strings.Join([]string{
		"NUMBER=1.25",
		"NAN=NaN",
		"BAD_NUMBER=one",
		"EMPTY=",
		"TEXT=value",
		"CHOICE=green",
		"LOW=1",
		"HIGH=10",
		"OUTSIDE=11",
	}, "\n"))

	if got, err := environment.Number("NUMBER"); err != nil || got != 1.25 {
		t.Fatalf("Number(NUMBER) = (%v, %v)", got, err)
	}
	if got, err := environment.Number("NAN"); err != nil || !math.IsNaN(got) {
		t.Fatalf("Number(NAN) = (%v, %v)", got, err)
	}
	if _, err := environment.Number("BAD_NUMBER"); err == nil || !strings.Contains(err.Error(), "invalid number") {
		t.Fatalf("Number(BAD_NUMBER) error = %v", err)
	}
	if _, err := environment.Number("MISSING"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("Number(MISSING) error = %v", err)
	}
	if got, err := environment.NumberOr("MISSING", 2.5); err != nil || got != 2.5 {
		t.Fatalf("NumberOr(MISSING) = (%v, %v)", got, err)
	}
	if got, err := environment.NumberOr("NUMBER", 2.5); err != nil || got != 1.25 {
		t.Fatalf("NumberOr(NUMBER) = (%v, %v)", got, err)
	}
	if _, err := environment.NumberOr("BAD_NUMBER", 2.5); err == nil {
		t.Fatalf("NumberOr(BAD_NUMBER) error = nil")
	}

	if got, err := environment.NonEmpty("TEXT"); err != nil || got != "value" {
		t.Fatalf("NonEmpty(TEXT) = (%q, %v)", got, err)
	}
	if _, err := environment.NonEmpty("EMPTY"); err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("NonEmpty(EMPTY) error = %v", err)
	}
	if _, err := environment.NonEmpty("MISSING"); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("NonEmpty(MISSING) error = %v", err)
	}

	if got, err := environment.Choice("CHOICE", []string{"red", "green", "blue"}); err != nil || got != "green" {
		t.Fatalf("Choice(CHOICE) = (%q, %v)", got, err)
	}
	if _, err := environment.Choice("CHOICE", []string{"red", "blue"}); err == nil || !strings.Contains(err.Error(), "expected one of red, blue") {
		t.Fatalf("Choice(disallowed) error = %v", err)
	}
	if _, err := environment.Choice("CHOICE", nil); err == nil || !strings.Contains(err.Error(), "must not be empty") {
		t.Fatalf("Choice(empty allowed) error = %v", err)
	}
	if _, err := environment.Choice("MISSING", []string{"value"}); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("Choice(MISSING) error = %v", err)
	}

	for _, name := range []string{"LOW", "HIGH"} {
		if _, err := environment.IntegerRange(name, 1, 10); err != nil {
			t.Fatalf("IntegerRange(%s) error = %v", name, err)
		}
	}
	if _, err := environment.IntegerRange("OUTSIDE", 1, 10); err == nil || !strings.Contains(err.Error(), "outside 1..10") {
		t.Fatalf("IntegerRange(OUTSIDE) error = %v", err)
	}
	if _, err := environment.IntegerRange("LOW", 10, 1); err == nil || !strings.Contains(err.Error(), "invalid integer range") {
		t.Fatalf("IntegerRange(invalid range) error = %v", err)
	}
	if _, err := environment.IntegerRange("MISSING", 1, 10); err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("IntegerRange(MISSING) error = %v", err)
	}
}

func TestRequire(t *testing.T) {
	environment := mustParse(t, "A=\nB=value\n")
	if err := environment.Require(nil); err != nil {
		t.Fatalf("Require(nil) error = %v", err)
	}
	if err := environment.Require([]string{"A", "B"}); err != nil {
		t.Fatalf("Require(present) error = %v", err)
	}
	if err := environment.Require([]string{"A", "MISSING", "B"}); err == nil || !strings.Contains(err.Error(), "MISSING") {
		t.Fatalf("Require(missing) error = %v", err)
	}
}

func TestStrictDuplicatePolicy(t *testing.T) {
	if _, err := parseStrict("A=first\nB=value\n"); err != nil {
		t.Fatalf("parseStrict(unique) error = %v", err)
	}
	if _, err := parseStrict("A=first\nA=second\n"); err == nil || !strings.Contains(err.Error(), "line 2: duplicate environment variable A") {
		t.Fatalf("parseStrict(duplicate) error = %v", err)
	}
	if got := mustParse(t, "A=first\nA=second\n").TextOr("A", ""); got != "second" {
		t.Fatalf("parse duplicate value = %q, want second", got)
	}
}

func TestFileLoadingVariants(t *testing.T) {
	directory := t.TempDir()
	validPath := filepath.Join(directory, "valid.env")
	duplicatePath := filepath.Join(directory, "duplicate.env")
	if err := os.WriteFile(validPath, []byte("A=value\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(duplicatePath, []byte("A=first\nA=second\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	if environment, err := loadStrict(validPath); err != nil || environment.TextOr("A", "") != "value" {
		t.Fatalf("loadStrict(valid) = (%#v, %v)", environment, err)
	}
	if _, err := loadStrict(duplicatePath); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("loadStrict(duplicate) error = %v", err)
	}
	if _, err := loadStrict(filepath.Join(directory, "missing")); err == nil {
		t.Fatalf("loadStrict(missing) error = nil")
	}

	if environment, err := loadIfExists(validPath); err != nil || environment.TextOr("A", "") != "value" {
		t.Fatalf("loadIfExists(valid) = (%#v, %v)", environment, err)
	}
	if environment, err := loadIfExists(filepath.Join(directory, "missing")); err != nil || environment.Size() != 0 {
		t.Fatalf("loadIfExists(missing) = (%#v, %v)", environment, err)
	}
	if _, err := loadIfExists(directory); err == nil {
		t.Fatalf("loadIfExists(directory) error = nil")
	}
}

func TestFromProcessAll(t *testing.T) {
	const key = "KINMOKUSEI_ENV_TEST_ALL"
	t.Setenv(key, "all-value")
	environment := fromProcessAll()
	if got := environment.TextOr(key, ""); got != "all-value" {
		t.Fatalf("fromProcessAll()[%s] = %q", key, got)
	}
}

func TestEncodeIsDeterministicAndRoundTrips(t *testing.T) {
	environment := NewEnvironment()
	values := map[string]string{
		"Z": "last",
		"A": "line one\nline two\r\ttab",
		"Q": `quote " and slash \\ and #`,
		"E": "",
	}
	for name, value := range values {
		if err := environment.Set(name, value); err != nil {
			t.Fatal(err)
		}
	}

	encoded := environment.Encode()
	if !strings.HasPrefix(encoded, "A=\"") || !strings.HasSuffix(encoded, "Z=\"last\"\n") {
		t.Fatalf("Encode() order = %q", encoded)
	}
	decoded := mustParse(t, encoded)
	for name, want := range values {
		if got := decoded.TextOr(name, "missing"); got != want {
			t.Errorf("round trip %s = %q, want %q", name, got, want)
		}
	}
}

func TestExpansionMatrix(t *testing.T) {
	environment := mustParse(t, strings.Join([]string{
		"HOST=localhost",
		"PORT=8080",
		"BASE=http://${HOST}:${PORT}",
		"API=${BASE}/api",
		"COPY=${HOST}-${HOST}",
		"LITERAL=$HOST",
	}, "\n"))
	expanded, err := environment.Expanded()
	if err != nil {
		t.Fatalf("Expanded() error = %v", err)
	}
	wants := map[string]string{
		"BASE":    "http://localhost:8080",
		"API":     "http://localhost:8080/api",
		"COPY":    "localhost-localhost",
		"LITERAL": "$HOST",
	}
	for name, want := range wants {
		if got := expanded.TextOr(name, ""); got != want {
			t.Errorf("Expanded().TextOr(%s) = %q, want %q", name, got, want)
		}
	}
	if environment.TextOr("BASE", "") != "http://${HOST}:${PORT}" {
		t.Fatalf("Expanded() mutated the source")
	}
}

func TestExpansionFailures(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"undefined", "A=${MISSING}", "references undefined variable MISSING"},
		{"unterminated", "A=${MISSING", "unterminated expansion"},
		{"empty name", "A=${}", `invalid expansion name ""`},
		{"invalid name", "A=${BAD-NAME}", `invalid expansion name "BAD-NAME"`},
		{"direct cycle", "A=${A}", "cyclic environment variable expansion involving A"},
		{"indirect cycle", "A=${B}\nB=${C}\nC=${A}", "cyclic environment variable expansion"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := mustParse(t, test.source).Expanded()
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Expanded() error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestResolveRejectsUnknownRoot(t *testing.T) {
	_, err := NewEnvironment().resolve("MISSING", map[string]int{}, NewEnvironment())
	if err == nil || !strings.Contains(err.Error(), "is not defined") {
		t.Fatalf("resolve(MISSING) error = %v", err)
	}
}

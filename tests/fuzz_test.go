package env

import "testing"

func FuzzParseNeverPanics(f *testing.F) {
	for _, source := range []string{
		"",
		"NAME=value",
		"NAME='value'",
		`NAME="value\n"`,
		"\ufeffNAME=value\r\n",
		"NAME=value # comment",
		"NAME=\xff",
	} {
		f.Add(source)
	}
	f.Fuzz(func(t *testing.T, source string) {
		_, _ = parse(source)
	})
}

func FuzzSetAndTypedAccessNeverPanic(f *testing.F) {
	f.Add("NAME", "1")
	f.Add("", "")
	f.Add("BAD-NAME", "true")
	f.Fuzz(func(t *testing.T, name string, value string) {
		environment := NewEnvironment()
		if environment.Set(name, value) == nil {
			_, _ = environment.Text(name)
			_, _ = environment.Integer(name)
			_, _ = environment.Boolean(name)
		}
	})
}

func FuzzEncodeRoundTrip(f *testing.F) {
	f.Add("plain")
	f.Add("line one\nline two")
	f.Add(`quote " slash \\ hash # tab\t`)
	f.Add("金木犀")
	f.Fuzz(func(t *testing.T, value string) {
		environment := NewEnvironment()
		if err := environment.Set("VALUE", value); err != nil {
			t.Fatal(err)
		}
		decoded, err := parse(environment.Encode())
		if err != nil {
			t.Fatalf("parse(Encode()) error = %v", err)
		}
		if got := decoded.TextOr("VALUE", "missing"); got != value {
			t.Fatalf("round trip = %q, want %q", got, value)
		}
	})
}

func FuzzExpansionNeverPanics(f *testing.F) {
	f.Add("literal")
	f.Add("${A}")
	f.Add("${MISSING}")
	f.Add("${")
	f.Fuzz(func(t *testing.T, value string) {
		environment := NewEnvironment()
		if environment.Set("A", value) == nil {
			_, _ = environment.Expanded()
		}
	})
}

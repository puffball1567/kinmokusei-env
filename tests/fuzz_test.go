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

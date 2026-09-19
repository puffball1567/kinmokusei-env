#!/bin/sh
set -eu

project_dir=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
test_dir=$(mktemp -d)
trap 'rm -rf "$test_dir"' EXIT HUP INT TERM

keika_command=${KEIKA:-keika}
go_cache="$test_dir/go-cache"
generated_dir="$test_dir/generated"
mkdir -p "$go_cache" "$generated_dir"

cd "$project_dir"
"$keika_command" deps lock
"$keika_command" check env.km
"$keika_command" check examples/basic/main.km
"$keika_command" emit-go -package env -o "$generated_dir/env.go" env.km
cp tests/*_test.go "$generated_dir/"

cd "$generated_dir"
go mod init example.com/kinmokusei-env-test >/dev/null
GOCACHE="$go_cache" go test -count=1 -coverprofile=coverage.out ./...
coverage_total=$(GOCACHE="$go_cache" go tool cover -func=coverage.out | awk '/^total:/ { print $3 }')
if [ "$coverage_total" != "100.0%" ]; then
  printf 'statement coverage is %s, want 100.0%%\n' "$coverage_total" >&2
  exit 1
fi
GOCACHE="$go_cache" go test -run='^$' -fuzz='^FuzzParseNeverPanics$' -fuzztime=1s
GOCACHE="$go_cache" go test -run='^$' -fuzz='^FuzzSetAndTypedAccessNeverPanic$' -fuzztime=1s

cd "$project_dir"
"$keika_command" build -o "$test_dir/example" examples/basic/main.km
mkdir "$test_dir/example-run"
printf 'HOST=127.0.0.1\nPORT=7000\nDEBUG=true\n' >"$test_dir/example-run/.env"
actual_output=$(cd "$test_dir/example-run" && "$test_dir/example")
expected_output='listening on 127.0.0.1:7000 (debug=true)'
if [ "$actual_output" != "$expected_output" ]; then
  printf 'unexpected example output:\n%s\n' "$actual_output" >&2
  exit 1
fi

library_dir="$test_dir/library"
consumer_dir="$test_dir/external-consumer"
mkdir "$library_dir"
cp env.km go.mod kinmokusei.lock kinmokusei.toml "$library_dir/"
"$keika_command" new app --module example.com/kinmokusei-env-consumer "$consumer_dir"
cp tests/external_consumer/main.km "$consumer_dir/main.km"
cd "$consumer_dir"
"$keika_command" deps add --offline --replace ../library github.com/puffball1567/kinmokusei-env@v0.1.0
"$keika_command" check
"$keika_command" build -o "$test_dir/external-consumer-app"
actual_output=$("$test_dir/external-consumer-app")
expected_output='localhost:9000 debug=true'
if [ "$actual_output" != "$expected_output" ]; then
  printf 'unexpected external-consumer output:\n%s\n' "$actual_output" >&2
  exit 1
fi

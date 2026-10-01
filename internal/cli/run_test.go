package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/cli"
)

const empty = `{"version":1,"mappings":[]}`
const added = `{"version":1,"mappings":[{"registry":"r","route":"x","origin":{"chain":"a","kind":"opaque","address":"a","decimals":1},"destination":{"chain":"b","kind":"opaque","address":"b","decimals":1}}]}`
const zero = `{"version":1,"mappings":[{"registry":"r","route":"x","origin":{"chain":"a","kind":"evm","address":"0x0000000000000000000000000000000000000000","decimals":1},"destination":{"chain":"b","kind":"opaque","address":"b","decimals":1}}]}`

func TestRun_returns_exact_exit_codes(t *testing.T) {
	tests := []struct {
		name   string
		before string
		after  string
		args   []string
		want   int
	}{
		{"no change", empty, empty, nil, 0},
		{"informational change", empty, added, nil, 1},
		{"review finding", empty, zero, nil, 2},
		{"usage error", empty, empty, []string{"only-one"}, 3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Given
			dir := t.TempDir()
			before := filepath.Join(dir, "before.json")
			after := filepath.Join(dir, "after.json")
			if err := os.WriteFile(before, []byte(test.before), 0o600); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(after, []byte(test.after), 0o600); err != nil {
				t.Fatal(err)
			}
			args := test.args
			if args == nil {
				args = []string{"--schema", "native", before, after}
			}

			// When
			var stdout, stderr bytes.Buffer
			got := cli.Run(args, &stdout, &stderr)

			// Then
			if got != test.want {
				t.Fatalf("exit = %d, want %d; stderr=%s", got, test.want, stderr.String())
			}
		})
	}
}

func TestRun_matches_JSON_and_text_goldens(t *testing.T) {
	for _, format := range []string{"json", "text"} {
		t.Run(format, func(t *testing.T) {
			// Given
			dir := t.TempDir()
			before := filepath.Join(dir, "before.json")
			after := filepath.Join(dir, "after.json")
			_ = os.WriteFile(before, []byte(empty), 0o600)
			_ = os.WriteFile(after, []byte(added), 0o600)
			want, err := os.ReadFile(filepath.Join("testdata", format+".golden"))
			if err != nil {
				t.Fatal(err)
			}

			// When
			var output bytes.Buffer
			cli.Run([]string{"--schema", "native", "--format", format, before, after}, &output, &bytes.Buffer{})

			// Then
			if !bytes.Equal(output.Bytes(), want) {
				t.Fatalf("output:\n%s\nwant:\n%s", output.Bytes(), want)
			}
		})
	}
}

func TestRun_requires_explicit_schema(t *testing.T) {
	// Given
	var stdout, stderr bytes.Buffer

	// When
	exit := cli.Run([]string{"before.json", "after.json"}, &stdout, &stderr)

	// Then
	if exit != 3 {
		t.Fatalf("exit = %d, want 3", exit)
	}
}

func TestRun_replays_real_uniswap_local_map_shape(t *testing.T) {
	// Given
	dir := t.TempDir()
	before := filepath.Join(dir, "before.json")
	after := filepath.Join(dir, "after.json")
	if err := os.WriteFile(before, []byte(`{}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(after, []byte(`{"0xdac17f958d2ee523a2206206994597c13d831ec7":{"childToken":"0xe7cd86e13AC4309349F30B3435a9d337750fC82D","decimals":6}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	// When
	var stdout, stderr bytes.Buffer
	exit := cli.Run([]string{"--schema", "uniswap-local-map", "--destination-chain", "monad", "--format", "json", before, after}, &stdout, &stderr)

	// Then
	if exit != 1 || !bytes.Contains(stdout.Bytes(), []byte(`"code":"addition"`)) {
		t.Fatalf("exit = %d, stdout = %s, stderr = %s", exit, stdout.String(), stderr.String())
	}
}

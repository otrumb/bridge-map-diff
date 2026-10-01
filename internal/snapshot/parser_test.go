package snapshot_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func TestParse_normalizes_valid_EVM_and_opaque_addresses(t *testing.T) {
	// Given
	raw := `{"version":1,"mappings":[{"registry":"r","scope":"","route":"canonical","origin":{"chain":"ethereum","kind":"evm","address":"0x52908400098527886E0F7030069857D2E4169EE7","decimals":18},"destination":{"chain":"solana","kind":"opaque","address":"  AbC123  ","decimals":8},"symbol":"TOK"}]}`

	// When
	got, err := snapshot.Parse([]byte(raw))

	// Then
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Mappings[0].Origin.Address != "0x52908400098527886e0f7030069857d2e4169ee7" {
		t.Fatalf("origin address = %q", got.Mappings[0].Origin.Address)
	}
	if got.Mappings[0].Destination.Address != "AbC123" {
		t.Fatalf("destination address = %q", got.Mappings[0].Destination.Address)
	}
}

func TestParse_rejects_invalid_mixed_case_EVM_checksum(t *testing.T) {
	// Given
	raw := `{"version":1,"mappings":[{"registry":"r","route":"x","origin":{"chain":"ethereum","kind":"evm","address":"0x52908400098527886E0F7030069857D2E4169Ee7","decimals":18},"destination":{"chain":"base","kind":"evm","address":"0xde709f2102306220921060314715629080e2fb77","decimals":18}}]}`

	// When
	_, err := snapshot.Parse([]byte(raw))

	// Then
	if !errors.Is(err, snapshot.ErrInvalidAddress) {
		t.Fatalf("error = %v, want ErrInvalidAddress", err)
	}
}

func TestParse_accepts_uniform_case_EVM_addresses(t *testing.T) {
	tests := []string{
		"0xde709f2102306220921060314715629080e2fb77",
		"0XDE709F2102306220921060314715629080E2FB77",
	}
	for _, address := range tests {
		t.Run(address, func(t *testing.T) {
			// Given
			raw := `{"version":1,"mappings":[{"registry":"r","route":"x","origin":{"chain":"a","kind":"evm","address":"` + address + `","decimals":18},"destination":{"chain":"b","kind":"opaque","address":"token","decimals":6}}]}`

			// When
			got, err := snapshot.Parse([]byte(raw))

			// Then
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.Mappings[0].Origin.Address != "0xde709f2102306220921060314715629080e2fb77" {
				t.Fatalf("address = %q", got.Mappings[0].Origin.Address)
			}
		})
	}
}

func TestParse_rejects_untrusted_input_classes(t *testing.T) {
	tests := []struct {
		name string
		raw  []byte
		err  error
	}{
		{"unknown field", []byte(`{"version":1,"mappings":[],"extra":true}`), snapshot.ErrInvalidSchema},
		{"duplicate root key", []byte(`{"version":1,"version":1,"mappings":[]}`), snapshot.ErrInvalidSchema},
		{"duplicate nested key", []byte(`{"version":1,"mappings":[{"registry":"r","registry":"r","route":"x","origin":{"chain":"a","kind":"opaque","address":"a","decimals":1},"destination":{"chain":"b","kind":"opaque","address":"b","decimals":1}}]}`), snapshot.ErrInvalidSchema},
		{"NUL", []byte("{\"version\":1,\"mappings\":[],\"x\":\"\x00\"}"), snapshot.ErrNUL},
		{"invalid UTF-8", []byte{0xff}, snapshot.ErrInvalidUTF8},
		{"depth over 64", []byte(strings.Repeat("[", 65) + strings.Repeat("]", 65)), snapshot.ErrDepthLimit},
		{"file over 16 MiB", make([]byte, 16<<20+1), snapshot.ErrSizeLimit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// When
			_, err := snapshot.Parse(test.raw)

			// Then
			if !errors.Is(err, test.err) {
				t.Fatalf("error = %v, want %v", err, test.err)
			}
		})
	}
}

func TestParse_rejects_identity_delimiter_in_identity_fields(t *testing.T) {
	// Given
	raw := `{"version":1,"mappings":[{"registry":"r|other","route":"x","origin":{"chain":"a","kind":"opaque","address":"a","decimals":1},"destination":{"chain":"b","kind":"opaque","address":"b","decimals":1}}]}`

	// When
	_, err := snapshot.Parse([]byte(raw))

	// Then
	if !errors.Is(err, snapshot.ErrInvalidSchema) {
		t.Fatalf("error = %v, want ErrInvalidSchema", err)
	}
}

func TestParse_rejects_more_than_100000_mappings(t *testing.T) {
	// Given
	mapping := `{"registry":"r","route":"x","origin":{"chain":"a","kind":"opaque","address":"a","decimals":1},"destination":{"chain":"b","kind":"opaque","address":"b","decimals":1}}`
	raw := []byte(`{"version":1,"mappings":[` + strings.Repeat(mapping+",", 100000) + mapping + `]}`)

	// When
	_, err := snapshot.Parse(raw)

	// Then
	if !errors.Is(err, snapshot.ErrMappingLimit) {
		t.Fatalf("error = %v, want ErrMappingLimit", err)
	}
}

func TestParse_preserves_unknown_and_known_zero_decimals(t *testing.T) {
	// Given
	raw := `{"version":1,"mappings":[{"registry":"r","route":"x","origin":{"chain":"a","kind":"opaque","address":"a"},"destination":{"chain":"b","kind":"opaque","address":"b","decimals":0}}]}`

	// When
	got, err := snapshot.Parse([]byte(raw))

	// Then
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Mappings[0].Origin.Decimals != nil {
		t.Fatalf("origin decimals = %v, want unknown", got.Mappings[0].Origin.Decimals)
	}
	if got.Mappings[0].Destination.Decimals == nil || *got.Mappings[0].Destination.Decimals != 0 {
		t.Fatalf("destination decimals = %v, want known zero", got.Mappings[0].Destination.Decimals)
	}
}

func TestParse_treats_null_decimals_as_unknown(t *testing.T) {
	// Given
	raw := `{"version":1,"mappings":[{"registry":"r","route":"x","origin":{"chain":"a","kind":"opaque","address":"a","decimals":null},"destination":{"chain":"b","kind":"opaque","address":"b"}}]}`

	// When
	got, err := snapshot.Parse([]byte(raw))

	// Then
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Mappings[0].Origin.Decimals != nil {
		t.Fatalf("origin decimals = %v, want unknown", got.Mappings[0].Origin.Decimals)
	}
}

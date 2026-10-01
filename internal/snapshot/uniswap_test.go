package snapshot_test

import (
	"errors"
	"testing"

	"github.com/otrumb/bridge-map-diff/internal/snapshot"
)

func TestParseUniswapLocalMap_parses_real_monad_shape(t *testing.T) {
	// Given
	raw := `{"0xdac17f958d2ee523a2206206994597c13d831ec7":{"childToken":"0xe7cd86e13AC4309349F30B3435a9d337750fC82D","decimals":6}}`

	// When
	got, err := snapshot.ParseUniswapLocalMap([]byte(raw), "monad")

	// Then
	if err != nil {
		t.Fatalf("ParseUniswapLocalMap() error = %v", err)
	}
	if len(got.Mappings) != 1 {
		t.Fatalf("mappings = %d, want 1", len(got.Mappings))
	}
	mapping := got.Mappings[0]
	if mapping.Registry != "uniswap" || mapping.Route != "local" || mapping.Origin.Chain != "ethereum" || mapping.Destination.Chain != "monad" {
		t.Fatalf("mapping context = %#v", mapping)
	}
	if mapping.Origin.Decimals == nil || *mapping.Origin.Decimals != 6 || mapping.Destination.Decimals == nil || *mapping.Destination.Decimals != 6 {
		t.Fatalf("decimals = %v/%v, want 6/6", mapping.Origin.Decimals, mapping.Destination.Decimals)
	}
}

func TestParseUniswapLocalMap_rejects_missing_chain_context(t *testing.T) {
	// Given
	raw := `{"0xdac17f958d2ee523a2206206994597c13d831ec7":{"childToken":"0xe7cd86e13AC4309349F30B3435a9d337750fC82D","decimals":6}}`

	// When
	_, err := snapshot.ParseUniswapLocalMap([]byte(raw), "")

	// Then
	if !errors.Is(err, snapshot.ErrInvalidSchema) {
		t.Fatalf("error = %v, want ErrInvalidSchema", err)
	}
}

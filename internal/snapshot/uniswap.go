package snapshot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"
)

type uniswapMappedToken struct {
	ChildToken string `json:"childToken"`
	Decimals   *uint8 `json:"decimals"`
}

func ParseUniswapLocalMap(data []byte, destinationChain string) (Snapshot, error) {
	destinationChain = strings.TrimSpace(destinationChain)
	if destinationChain == "" || len(data) > maxFileSize || !utf8.Valid(data) || bytes.IndexByte(data, 0) >= 0 || exceedsDepth(data, 64) {
		return Snapshot{}, ErrInvalidSchema
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return Snapshot{}, err
	}
	var source map[string]uniswapMappedToken
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&source); err != nil || source == nil {
		return Snapshot{}, ErrInvalidSchema
	}
	if err := rejectTrailing(decoder); err != nil {
		return Snapshot{}, err
	}
	origins := make([]string, 0, len(source))
	for origin := range source {
		origins = append(origins, origin)
	}
	slices.Sort(origins)
	result := Snapshot{Version: 1, Mappings: make([]Mapping, 0, len(origins))}
	for _, origin := range origins {
		token := source[origin]
		mapping := Mapping{
			Registry: "uniswap", Route: "local",
			Origin:      Endpoint{Chain: "ethereum", Kind: "evm", Address: origin, Decimals: token.Decimals},
			Destination: Endpoint{Chain: destinationChain, Kind: "evm", Address: token.ChildToken, Decimals: token.Decimals},
		}
		if token.Decimals == nil {
			return Snapshot{}, fmt.Errorf("origin %s: %w", origin, ErrInvalidSchema)
		}
		if err := normalizeMapping(&mapping); err != nil {
			return Snapshot{}, fmt.Errorf("origin %s: %w", origin, err)
		}
		result.Mappings = append(result.Mappings, mapping)
	}
	return result, nil
}

package snapshot

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/sha3"
)

const maxFileSize = 16 << 20

func Parse(data []byte) (Snapshot, error) {
	if len(data) > maxFileSize {
		return Snapshot{}, ErrSizeLimit
	}
	if !utf8.Valid(data) {
		return Snapshot{}, ErrInvalidUTF8
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return Snapshot{}, ErrNUL
	}
	if exceedsDepth(data, 64) {
		return Snapshot{}, ErrDepthLimit
	}
	if err := rejectDuplicateKeys(data); err != nil {
		return Snapshot{}, err
	}

	var parsed Snapshot
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&parsed); err != nil {
		return Snapshot{}, fmt.Errorf("%w: %v", ErrInvalidSchema, err)
	}
	if err := rejectTrailing(decoder); err != nil {
		return Snapshot{}, err
	}
	if parsed.Version != 1 || parsed.Mappings == nil {
		return Snapshot{}, ErrInvalidSchema
	}
	if len(parsed.Mappings) > 100000 {
		return Snapshot{}, ErrMappingLimit
	}
	for index := range parsed.Mappings {
		if err := normalizeMapping(&parsed.Mappings[index]); err != nil {
			return Snapshot{}, fmt.Errorf("mapping %d: %w", index, err)
		}
	}
	return parsed, nil
}

func rejectTrailing(decoder *json.Decoder) error {
	var extra json.RawMessage
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return ErrInvalidSchema
}

func rejectDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	if err := scanJSONValue(decoder); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidSchema, err)
	}
	return rejectTrailing(decoder)
}

func scanJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, compound := token.(json.Delim)
	if !compound {
		return nil
	}
	switch delimiter {
	case '{':
		keys := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok || keys[key] {
				return errors.New("duplicate or invalid object key")
			}
			keys[key] = true
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err := scanJSONValue(decoder); err != nil {
				return err
			}
		}
	default:
		return errors.New("invalid JSON delimiter")
	}
	_, err = decoder.Token()
	return err
}

func normalizeMapping(mapping *Mapping) error {
	fields := []*string{&mapping.Registry, &mapping.Scope, &mapping.Route, &mapping.Symbol, &mapping.Supersedes,
		&mapping.Origin.Chain, &mapping.Origin.Kind, &mapping.Origin.Address,
		&mapping.Destination.Chain, &mapping.Destination.Kind, &mapping.Destination.Address}
	for _, field := range fields {
		*field = strings.TrimSpace(*field)
		if len(*field) > 1024 {
			return ErrStringLimit
		}
	}
	if mapping.Registry == "" || mapping.Route == "" || mapping.Origin.Chain == "" || mapping.Destination.Chain == "" {
		return ErrInvalidSchema
	}
	identityFields := []string{mapping.Registry, mapping.Scope, mapping.Route, mapping.Origin.Chain, mapping.Origin.Address, mapping.Destination.Chain}
	for _, field := range identityFields {
		if strings.ContainsRune(field, '|') {
			return ErrInvalidSchema
		}
	}
	var err error
	mapping.Origin.Address, err = normalizeAddress(mapping.Origin.Kind, mapping.Origin.Address)
	if err != nil {
		return fmt.Errorf("origin: %w", err)
	}
	mapping.Destination.Address, err = normalizeAddress(mapping.Destination.Kind, mapping.Destination.Address)
	if err != nil {
		return fmt.Errorf("destination: %w", err)
	}
	return nil
}

func normalizeAddress(kind, address string) (string, error) {
	switch kind {
	case "opaque":
		if address == "" {
			return "", ErrInvalidAddress
		}
		return address, nil
	case "evm":
		return normalizeEVM(address)
	default:
		return "", ErrInvalidSchema
	}
}

func normalizeEVM(address string) (string, error) {
	if len(address) != 42 || (address[:2] != "0x" && address[:2] != "0X") {
		return "", ErrInvalidAddress
	}
	hexAddress := address[2:]
	if _, err := hex.DecodeString(hexAddress); err != nil {
		return "", ErrInvalidAddress
	}
	if hexAddress != strings.ToLower(hexAddress) && hexAddress != strings.ToUpper(hexAddress) && !validChecksum(hexAddress) {
		return "", ErrInvalidAddress
	}
	return "0x" + strings.ToLower(hexAddress), nil
}

func validChecksum(address string) bool {
	lower := strings.ToLower(address)
	hash := sha3.NewLegacyKeccak256()
	_, _ = hash.Write([]byte(lower))
	digest := hash.Sum(nil)
	for index, character := range address {
		if character < 'a' || character > 'f' {
			if character < 'A' || character > 'F' {
				continue
			}
		}
		nibble := digest[index/2]
		if index%2 == 0 {
			nibble >>= 4
		} else {
			nibble &= 0x0f
		}
		if (nibble >= 8) != (character >= 'A' && character <= 'F') {
			return false
		}
	}
	return true
}

func exceedsDepth(data []byte, limit int) bool {
	depth := 0
	inString := false
	escaped := false
	for _, character := range data {
		if inString {
			if escaped {
				escaped = false
			} else if character == '\\' {
				escaped = true
			} else if character == '"' {
				inString = false
			}
			continue
		}
		switch character {
		case '"':
			inString = true
		case '{', '[':
			depth++
			if depth > limit {
				return true
			}
		case '}', ']':
			depth--
		}
	}
	return false
}

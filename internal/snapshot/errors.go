package snapshot

import "errors"

var (
	ErrSizeLimit      = errors.New("snapshot exceeds 16 MiB")
	ErrDepthLimit     = errors.New("snapshot exceeds depth 64")
	ErrMappingLimit   = errors.New("snapshot exceeds 100000 mappings")
	ErrStringLimit    = errors.New("snapshot string exceeds 1024 bytes")
	ErrInvalidUTF8    = errors.New("snapshot contains invalid UTF-8")
	ErrNUL            = errors.New("snapshot contains NUL")
	ErrInvalidSchema  = errors.New("invalid native snapshot schema")
	ErrInvalidAddress = errors.New("invalid endpoint address")
)

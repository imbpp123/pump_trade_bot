package types

import "errors"

var (
	ErrParseDirectionNotFound = errors.New("signal direction not found")
	ErrParseSymbolNotFound    = errors.New("signal symbol not found")
	ErrParseLeverageNotFound  = errors.New("signal leverage not found")
	ErrParseEntryNotFound     = errors.New("signal entry not found")
	ErrParseTargetNotFound    = errors.New("signal target not found")
	ErrParseStopNotFound      = errors.New("signal stop not found")
)

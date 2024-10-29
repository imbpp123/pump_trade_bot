package types

import "errors"

var (
	ErrParsePositionNotFound         = errors.New("signal position not found")
	ErrParseSymbolNotFound           = errors.New("signal symbol not found")
	ErrParseLeverageIntervalNotFound = errors.New("signal leverage not found")
	ErrParseEntryIntervalNotFound    = errors.New("signal entry not found")
	ErrParseTargetNotFound           = errors.New("signal target not found")
	ErrParseStopNotFound             = errors.New("signal stop not found")
	ErrSignalHandlerNotFound         = errors.New("signal handler not found")
)

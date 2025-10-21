package domain

import (
	"errors"
)

// Common errors used throughout the plugin system
var (
	ErrNotFound          = errors.New("entity not found")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrAlreadyExists     = errors.New("already exists")
	ErrNotImplemented    = errors.New("not implemented")
	ErrInternal          = errors.New("internal error")
	ErrReadOnly          = errors.New("entity is read-only")
)

// All plugin-related interfaces are now defined in pkg/pluginsdk for SDK compatibility.
// Plugins use the SDK interfaces from pkg/pluginsdk (Plugin, IEntityProvider, etc.).
// The domain layer imports and uses SDK types through the pluginsdk package.

package common

import "errors"

// ErrNotFound this indicates that a record queried for in the database was not found
var ErrNotFound = errors.New("not found")

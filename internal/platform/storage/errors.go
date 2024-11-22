package storage

import "errors"

// ErrNotFound is an error to sql.ErrNoRows indicating that there was/were row(s) found.
var ErrNotFound = errors.New("not found")

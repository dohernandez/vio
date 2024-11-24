package database

import (
	"errors"
)

// ErrAlreadyExists error when entity already exists.
var ErrAlreadyExists = errors.New("entity already exists")

// ErrNotFound is an error to sql.ErrNoRows indicating that there was/were row(s) found.
var ErrNotFound = errors.New("not found")

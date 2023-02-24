// Package shortenerrors declares custom application errors.
package shortenerrors

import "errors"

// ErrItemAlreadyExists custom error for Already exist.
var ErrItemAlreadyExists = errors.New("item already exists in the database")

// ErrItemNotFound custom error for 404.
var ErrItemNotFound = errors.New("no url found by id")

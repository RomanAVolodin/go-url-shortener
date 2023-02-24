// Package shortenerrors declares custom application errors.
package shortenerrors

import "errors"

var ErrItemAlreadyExists = errors.New("item already exists in the database")
var ErrItemNotFound = errors.New("no url found by id")

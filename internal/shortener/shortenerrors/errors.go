package shortenerrors

import "errors"

var ErrItemAlreadyExists = errors.New("item already exists in the database")

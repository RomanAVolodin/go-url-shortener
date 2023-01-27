package shortenerrors

import "errors"

var ItemAlreadyExistsError = errors.New("Item already exists in the database")

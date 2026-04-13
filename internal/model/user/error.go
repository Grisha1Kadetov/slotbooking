package user

import "errors"

var ErrUnicEmail = errors.New("email must be unique")
var ErrNotFound = errors.New("user not found")
var ErrIncorrectPassword = errors.New("incorrect password")

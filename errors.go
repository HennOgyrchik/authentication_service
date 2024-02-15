package main

import "errors"

var ErrExpTimeHasExpired = errors.New("the expiration date has expired")
var InternalServerError = errors.New("internal server error")
var ErrNotFound = errors.New("not found")

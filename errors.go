package main

import "errors"

var ErrAlreadyExists = errors.New("The token already exists")
var ErrExpTimeHasExpired = errors.New("The expiration date has expired")
var InternalServerError = errors.New("Internal Server Error")
var ErrNotFound = errors.New("Not Found")

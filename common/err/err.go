package err

import "errors"

var ErrEventNotFound = errors.New("event not found")

var ErrEventExists = errors.New("event exists")
